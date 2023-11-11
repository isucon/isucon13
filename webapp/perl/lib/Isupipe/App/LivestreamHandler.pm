package Isupipe::App::LivestreamHandler;
use v5.38;
use utf8;
use experimental qw(try);

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::LivestreamViewer;
use Isupipe::Entity::Livestream;
use Isupipe::Entity::LivestreamTag;
use Isupipe::Entity::LivecommentReport;
use Isupipe::Entity::ReservationSlot;
use Isupipe::App::Util qw(
    check_params
    verify_user_session
    DEFAULT_USER_ID_KEY
);

use Isupipe::App::FillResponse qw(
    fill_livestream_response
);

use constant NUM_RESERVATION_SLOT => $ENV{ISUCON13_NUM_RESERVATION_SLOT} // 2;

use constant TERM_START_AT => 1711929600; # 2024/04/01 UTC
use constant TERM_END_AT   => 1743465600; # 2025/04/01 UTC;

use constant ReserveLivestreamRequest => Dict[
    tags          => ArrayRef[Int],
    title         => Str,
    description   => Str,
    # NOTE: コラボ配信の際に便利な自動スケジュールチェック機能
    # DBに記録しないが、コラボレーターがスケジュール的に問題ないか調べて、エラーを返す
    collaborators => ArrayRef[Int],
    start_at      => Int,
    end_at        => Int,
];

sub reserve_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};
    unless ($user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, "failed to find user-id from session");
    }

    my $params = $c->req->json_parameters;
    unless (check_params($params, ReserveLivestreamRequest)) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad request");
    }

    my $txn = $app->dbh->txn_scope;
    try {
        # 2024/04/01からの１年間の期間内であるかチェック
        infof('check term');
        if (!($params->{end_at} == TERM_END_AT || $params->{end_at} < TERM_END_AT) && (TERM_START_AT == $params->{start_at} || $params->{start_at} > TERM_START_AT)) {
            $c->halt_text(HTTP_BAD_REQUEST, "bad reservation time range");
        }

        infof('check collaborators');
        # 各ユーザについて、予約時間帯とかぶるような予約が存在しないか調べる (ある人は同時に複数の配信に物理的に出れない)
        my $user_ids = [];
        push @$user_ids, $user_id;
        push @$user_ids, $params->{collaborators}->@*;
        for my $user_id ($user_ids->@*) {
            my $founds = $app->dbh->select_one(
                'SELECT COUNT(*) FROM livestreams WHERE user_id = ? AND  ? >= start_at && ? <= end_at',
                $user_id,
                $params->{start_at},
                $params->{end_at},
            );
            if ($founds >= NUM_RESERVATION_SLOT) {
                $c->halt_text(HTTP_BAD_REQUEST, sprintf('ユーザ%dが予約できません', $user_id));
            }
        }

        # 予約枠をみて、予約が可能か調べる
        my $slots = $app->dbh->select_all_as(
            'Isupipe::Entity::ReservationSlot',
            'SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ?',
            $params->{start_at},
            $params->{end_at},
        );

        for my $slot ($slots->@*) {
            my $count = $app->dbh->select_one(
                'SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?',
                $slot->start_at,
                $slot->end_at,
            );
            infof('%d ~ %d予約枠の残数 = %d', $slot->start_at, $slot->end_at, $slot->slot);
            if ($count < 1) {
                $c->halt_text(HTTP_BAD_REQUEST, sprintf('予約区間 %d ~ %dが予約できません', $params->{start_at}, $params->{end_at}));
            }
        }

        my $livestream = Isupipe::Entity::Livestream->new(
            user_id       => $user_id,
            title         => $params->{title},
            description   => $params->{description},
            playlist_url  => "https://d2jpkt808jogxx.cloudfront.net/BigBuckBunny/playlist.m3u8",
            thumbnail_url => "https://picsum.photos/200/300",
            start_at      => $params->{start_at},
            end_at        => $params->{end_at},
        );

        infof('insert reservation slot');
        $app->dbh->query(
            'UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?',
            $params->{start_at},
            $params->{end_at},
        );

        infof('insert livestream');
        $app->dbh->query(
            'INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at)',
            $livestream->as_hashref,
        );

        infof('get inserted id');
        my $livestream_id = $app->dbh->last_insert_id;
        $livestream->id($livestream_id);

        infof('insert tags');
        # タグ追加
        for my $tag_id ($params->{tags}->@*) {
            $app->dbh->query(
                'INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (:livestream_id, :tag_id)',
                {
                    livestream_id => $livestream_id,
                    tag_id => $tag_id,
                },
            );
        }

        $livestream = fill_livestream_response($app, $livestream);

        $txn->commit;

        my $res = $c->render_json($livestream);
        $res->status(HTTP_CREATED);
        return $res;
    }
    catch ($e) {
        $txn->rollback;
        if ($e isa Kossy::Exception) {
            die $e;
        }
        $c->halt_text(HTTP_INTERNAL_SERVER_ERROR, $e);
    }
}

sub search_livestreams_handler($app, $c) {
    my $key_tag_name = $c->req->query_parameters->{tag};

    my $livestreams = [];
    if ($key_tag_name) {
        # タグによる取得
        my $tags = $app->dbh->select_all_as(
            'Isupipe::Entity::Tag',
            'SELECT id FROM tags WHERE name = ?',
            $key_tag_name
        );

        my $livestream_tags = $app->dbh->select_all_as(
            'Isupipe::Entity::LivestreamTag',
            'SELECT * FROM livestream_tags WHERE tag_id IN (?)',
            [map { $_->id } $tags->@*]
        );

        for my $livestream_tag ($livestream_tags->@*) {
            my $livestream = $app->dbh->select_row_as(
                'Isupipe::Entity::Livestream',
                'SELECT * FROM livestreams WHERE id = ?',
                $livestream_tag->livestream_id,
            );
            push $livestreams->@*, $livestream;
        }
    }
    else {
        # 検索条件なし
        my $query = 'SELECT * FROM livestreams';
        if ($c->req->query_parameters->{limit}) {
            $query .= ' LIMIT ?';
        }

        $livestreams = $app->dbh->select_all_as(
            'Isupipe::Entity::Livestream',
            $query,
            $c->req->query_parameters->{limit} // (),
        );
    }

    my $response = [
        map { fill_livestream_response($app, $_) } $livestreams->@*
    ];

    return $c->render_json($response);
}


sub get_my_livestreams_handler($app, $c) {
    verify_user_session($app, $c);

    my $txn = $app->dbh->txn_scope;

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $livestreams = $app->dbh->select_all_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE user_id = ?',
        $user_id,
    );

    my $response = [
        map { fill_livestream_response($app, $_) } $livestreams->@*
    ];

    $txn->commit;

    return $c->render_json($response);
}

sub get_user_livestreams_handler($app, $c) {
    verify_user_session($app, $c);

    my $username = $c->args->{username};

    my $txn = $app->dbh->txn_scope;
    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = ?',
        $username,
    );
    unless ($user) {
        $c->halt_text(HTTP_NOT_FOUND, "failed to get user");
    }

    my $livestreams = $app->dbh->select_all_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE user_id = ?',
        $user->id,
    );

    my $response = [
        map { fill_livestream_response($app, $_) } $livestreams->@*
    ];

    $txn->commit;

    return $c->render_json($response);
}

# FIXME: livestreamのカラムを追加し、視聴者数を増やす
#
# viewerテーブルの廃止
sub enter_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    my $user_id = $c->req->session->get(DEFAULT_USER_ID_KEY);
    unless ($user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, "failed to find user-id from session");
    }

    my $livesream_id = $c->args->{livestream_id};

    my $dbh = $app->dbh;
    my $txn = $dbh->txn_scope;

    my $viewer = Isupipe::Entity::LivestreamViewer->new(
        user_id => $user_id,
        livestream_id => $livesream_id,
    );

    $dbh->query(
        'INSERT INTO livestream_viewers (user_id, livestream_id) VALUES (:user_id, :livestream_id)',
        $viewer->as_hashref,
    );

    $dbh->query(
        'UPDATE livestreams SET viewer_count = viewer_count + 1 WHERE id = ?',
        $livesream_id,
    );

    $txn->commit;

    # FIXME: GO実装を返しているものが違う
    return $c->halt_no_content(HTTP_OK);
}

sub leave_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    my $live_stream_id = $c->args->{livestream_id};

    my $dbh = $app->dbh;
    my $txn = $dbh->txn_scope;

    $dbh->query(
        'UPDATE livestreams SET viewer_count = viewer_count - 1 WHERE id = ?',
        $live_stream_id,
    );

    $txn->commit;

    # FIXME: GO実装を返しているものが違う
    return $c->halt_no_content(HTTP_OK);
}

sub get_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    my $txn = $app->dbh->txn_scope;

    my $livestream = $app->dbh->select_row_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE id = ?',
        $livestream_id,
    );
    unless ($livestream) {
        $c->halt_text(HTTP_NOT_FOUND, "livestream not found");
    }

    my $response = fill_livestream_response($app, $livestream);

    $txn->commit;

    return $c->render_json($response);
}

sub get_livecomment_report_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    my $reports = $app->dbh->select_all_as(
        'Isupipe::Entity::LivecommentReport',
        'SELECT * FROM livecomment_reports WHERE livestream_id = ?',
        $livestream_id,
    );

    return $c->render_json($reports);
}


