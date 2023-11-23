package Isupipe::Handler::LivestreamHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::LivestreamViewer;
use Isupipe::Entity::Livestream;
use Isupipe::Entity::LivestreamTag;
use Isupipe::Entity::LivecommentReport;
use Isupipe::Entity::ReservationSlot;
use Isupipe::Util qw(
    check_params
    verify_user_session
    DEFAULT_USER_ID_KEY
);

use Isupipe::FillResponse qw(
    fill_livestream_response
    fill_livecomment_report_response
);

use constant NUM_RESERVATION_SLOT => $ENV{ISUCON13_NUM_RESERVATION_SLOT} // 2;

use constant TERM_START_AT => 1700874000; # 2023-11-25 10:00:00 JST
use constant TERM_END_AT   => 1732496400; # 2024-11-25 10:00:00 JST

use constant ReserveLivestreamRequest => Dict[
    tags          => ArrayRef[Int],
    title         => Str,
    description   => Str,
    playlist_url  => Str,
    thumbnail_url => Str,
    start_at      => Int,
    end_at        => Int,
];

sub reserve_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $params = $c->req->json_parameters;
    unless (check_params($params, ReserveLivestreamRequest)) {
        $c->halt(HTTP_BAD_REQUEST, "failed to decode the request body as jso");
    }

    my $txn = $app->dbh->txn_scope;

    # 2023/11/25 10:00からの１年間の期間内であるかチェック
    if (
        ($params->{start_at} == TERM_END_AT || $params->{start_at} > TERM_END_AT) ||
        ($params->{end_at} == TERM_START_AT || $params->{end_at} < TERM_START_AT)
    ) {
        $c->halt(HTTP_BAD_REQUEST, "bad reservation time range");
    }

    # 予約枠をみて、予約が可能か調べる
    # NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
    my $slots = $app->dbh->select_all_as(
        'Isupipe::Entity::ReservationSlot',
        'SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE',
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
            $c->halt(HTTP_BAD_REQUEST, sprintf("予約期間 %d ~ %dに対して、予約区間 %d ~ %dが予約できません", TERM_START_AT, TERM_END_AT, $params->{start_at}, $params->{end_at}));
        }
    }

    my $livestream = Isupipe::Entity::Livestream->new(
        user_id       => $user_id,
        title         => $params->{title},
        description   => $params->{description},
        playlist_url  => $params->{playlist_url},
        thumbnail_url => $params->{thumbnail_url},
        start_at      => $params->{start_at},
        end_at        => $params->{end_at},
    );

    $app->dbh->query(
        'UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?',
        $params->{start_at},
        $params->{end_at},
    );

    $app->dbh->query(
        'INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at)',
        $livestream->as_hashref,
    );

    my $livestream_id = $app->dbh->last_insert_id;
    $livestream->id($livestream_id);

    # タグ追加
    for my $tag_id ($params->{tags}->@*) {
        $app->dbh->query(
            'INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (:livestream_id, :tag_id)',
            {
                livestream_id => $livestream_id,
                tag_id        => $tag_id,
            },
        );
    }

    $livestream = fill_livestream_response($app, $livestream);

    $txn->commit;

    my $res = $c->render_json($livestream);
    $res->status(HTTP_CREATED);
    return $res;
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
            'SELECT * FROM livestream_tags WHERE tag_id IN (?) ORDER BY livestream_id DESC',
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
        my $query = 'SELECT * FROM livestreams ORDER BY id DESC';
        if (my $limit = $c->req->query_parameters->{limit}) {
            unless ($limit =~ /^\d+$/) {
                $c->halt(HTTP_BAD_REQUEST, "limit query parameter must be integer");
            }
            $query .= sprintf(" LIMIT %d", $limit);
        }

        $livestreams = $app->dbh->select_all_as(
            'Isupipe::Entity::Livestream',
            $query,
        );
    }

    $livestreams = [
        map { fill_livestream_response($app, $_) } $livestreams->@*
    ];

    return $c->render_json($livestreams);
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

    $livestreams = [
        map { fill_livestream_response($app, $_) } $livestreams->@*
    ];

    $txn->commit;

    return $c->render_json($livestreams);
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
        $c->halt(HTTP_NOT_FOUND, "user not found");
    }

    my $livestreams = $app->dbh->select_all_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE user_id = ?',
        $user->id,
    );

    $livestreams = [
        map { fill_livestream_response($app, $_) } $livestreams->@*
    ];

    $txn->commit;

    return $c->render_json($livestreams);
}

# viewerテーブルの廃止
sub enter_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $livesream_id = $c->args->{livestream_id};

    my $txn = $app->dbh->txn_scope;

    my $viewer = Isupipe::Entity::LivestreamViewer->new(
        user_id       => $user_id,
        livestream_id => $livesream_id,
        created_at    => time,
    );

    $app->dbh->query(
        "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(:user_id, :livestream_id, :created_at)",
        $viewer->as_hashref,
    );

    $txn->commit;

    return $c->halt_no_content(HTTP_OK);
}

sub exit_livestream_handler($app, $c) {
    verify_user_session($app, $c);

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $livestream_id = $c->args->{livestream_id};

    my $txn = $app->dbh->txn_scope;

    $app->dbh->query(
        "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?",
        $user_id,
        $livestream_id,
    );

    $txn->commit;

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
        $c->halt(HTTP_NOT_FOUND, "not found livestream that has the given id");
    }

    $livestream = fill_livestream_response($app, $livestream);

    $txn->commit;

    return $c->render_json($livestream);
}

sub get_livecomment_reports_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    my $txn = $app->dbh->txn_scope;

    my $livestream = $app->dbh->select_row_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE id = ?',
        $livestream_id,
    );
    unless ($livestream) {
        $c->halt(HTTP_INTERNAL_SERVER_ERROR, "failed to get livestream");
    }

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    if ($livestream->user_id != $user_id) {
        $c->halt(HTTP_FORBIDDEN, "can't get other streamer's livecomment reports");
    }

    my $reports = $app->dbh->select_all_as(
        'Isupipe::Entity::LivecommentReport',
        'SELECT * FROM livecomment_reports WHERE livestream_id = ?',
        $livestream_id,
    );

    $reports = [
        map { fill_livecomment_report_response($app, $_) } $reports->@*
    ];

    $txn->commit;

    return $c->render_json($reports);
}

