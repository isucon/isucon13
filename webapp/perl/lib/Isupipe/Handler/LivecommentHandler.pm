package Isupipe::Handler::LivecommentHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::Livecomment;
use Isupipe::Entity::LivecommentReport;
use Isupipe::Entity::NGWord;
use Isupipe::Util qw(
    verify_user_session
    DEFAULT_USER_ID_KEY
    check_params
);

use Isupipe::FillResponse qw(
    fill_livecomment_response
    fill_livecomment_report_response
);

use constant PostLivecommentRequest => Dict[
    comment => Str,
    tip => Int,
];

use constant ModerateRequest => Dict[
    ng_word => Str,
];

sub get_livecomments_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    my $txn = $app->dbh->txn_scope;

    my $query = "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC";
    if (my $limit = $c->req->query_parameters->{limit}) {
        unless ($limit =~ /^\d+$/) {
            $c->halt(HTTP_BAD_REQUEST, "limit query parameter must be integer");
        }
        $query .= sprintf(" LIMIT %d", $limit);
    }

    my $livecomments = $app->dbh->select_all_as(
        'Isupipe::Entity::Livecomment',
        $query,
        $livestream_id,
    );
    unless ($livecomments->@*) {
        $txn->rollback;
        return $c->render_json([]);
    }

    $livecomments = [
        map { fill_livecomment_response($app, $_) } $livecomments->@*
    ];

    $txn->commit;

    return $c->render_json($livecomments);
}

sub post_livecomment_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $params = $c->req->json_parameters;
    unless (check_params($params, PostLivecommentRequest)) {
        $c->halt(HTTP_BAD_REQUEST, "failed to decode the request body as json");
    }

    my $txn = $app->dbh->txn_scope;

    my $livestream = $app->dbh->select_row_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE id = ?',
        $livestream_id,
    );
    unless ($livestream) {
        $c->halt(HTTP_NOT_FOUND, "livestream not found");
    }

    # スパム判定
    my $ng_words = $app->dbh->select_all_as(
        'Isupipe::Entity::NGWord',
        'SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = ? AND livestream_id = ?',
        $livestream->user_id,
        $livestream->id,
    );

    my $hit_spam = 0;
    for my $ng_word ($ng_words->@*) {
        my $query =<<~ 'SQL';
            SELECT COUNT(*)
            FROM
            (SELECT ? AS text) AS texts
            INNER JOIN
            (SELECT CONCAT('%', ?, '%') AS pattern) AS patterns
            ON texts.text LIKE patterns.pattern
        SQL

        $hit_spam = $app->dbh->select_one($query, $params->{comment}, $ng_word->word);
        infof("[hitSpam=%d] comment=%s", $hit_spam, $params->{comment});
        if ($hit_spam >= 1) {
            $c->halt(HTTP_BAD_REQUEST, "このコメントがスパム判定されました");
        }
    }

    my $now = time;

    my $livecomment = Isupipe::Entity::Livecomment->new(
        user_id       => $user_id,
        livestream_id => $livestream_id,
        comment       => $params->{comment},
        tip           => $params->{tip},
        created_at    => $now,
    );

    $app->dbh->query(
        'INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (:user_id, :livestream_id, :comment, :tip, :created_at)',
        $livecomment->as_hashref,
    );

    my $livecomment_id = $app->dbh->last_insert_id;
    $livecomment->id($livecomment_id);

    $livecomment = fill_livecomment_response($app, $livecomment);

    $txn->commit;

    my $res = $c->render_json($livecomment);
    $res->status(HTTP_CREATED);
    return $res;
}

sub get_ngwords_handler($app, $c) {
    verify_user_session($app, $c);

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $livestream_id = $c->args->{livestream_id};

    my $ng_words = $app->dbh->select_all_as(
        'Isupipe::Entity::NGWord',
        'SELECT * FROM ng_words WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC',
        $user_id,
        $livestream_id,
    );

    return $c->render_json($ng_words);
}

sub report_livecomment_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};
    my $livecomment_id = $c->args->{livecomment_id};

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $txn = $app->dbh->txn_scope;

    my $livestream = $app->dbh->select_row_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE id = ?',
        $livestream_id,
    );
    unless ($livestream) {
        $c->halt(HTTP_NOT_FOUND, "livestream not found");
    }

    my $livecomment = $app->dbh->select_row_as(
        'Isupipe::Entity::Livecomment',
        'SELECT * FROM livecomments WHERE id = ?',
        $livecomment_id,
    );
    unless ($livecomment) {
        $c->halt(HTTP_NOT_FOUND, "livecomment not found");
    }

    my $report = Isupipe::Entity::LivecommentReport->new(
        user_id        => $user_id,
        livestream_id  => $livestream_id,
        livecomment_id => $livecomment_id,
        created_at     => time,
    );
    $app->dbh->query(
        'INSERT INTO livecomment_reports (user_id, livestream_id, livecomment_id, created_at) VALUES (:user_id, :livestream_id, :livecomment_id, :created_at)',
        $report->as_hashref,
    );
    my $report_id = $app->dbh->last_insert_id;
    $report->id($report_id);

    $report = fill_livecomment_report_response($app, $report);

    $txn->commit;

    my $res = $c->render_json($report);
    $res->status(HTTP_CREATED);
    return $res;
}

# NGワードを登録
sub moderate_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $params = $c->req->json_parameters;
    unless (check_params($params, ModerateRequest)) {
        $c->halt(HTTP_BAD_REQUEST, "bad request");
    }

    my $txn = $app->dbh->txn_scope;
    # 配信者自身の配信に対するmoderateなのかを検証
    my $owned_livestreams = $app->dbh->select_all(
        'SELECT * FROM livestreams WHERE id = ? AND user_id = ?',
        $livestream_id,
        $user_id,
    );
    if (@$owned_livestreams == 0) {
        $c->halt(HTTP_BAD_REQUEST, "A streamer can't moderate livestreams that other streamers own");
    }

    $app->dbh->query(
        'INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (:user_id, :livestream_id, :word, :created_at)',
        {
            user_id       => $user_id,
            livestream_id => $livestream_id,
            word          => $params->{ng_word},
            created_at    => time,
        },
    );

    my $word_id = $app->dbh->last_insert_id;

    my $ng_words = $app->dbh->select_all_as(
        'Isupipe::Entity::NGWord',
        'SELECT * FROM ng_words WHERE livestream_id = ?',
        $livestream_id,
    );

    # NGワードにヒットする過去の投稿も全削除する
    for my $ng_word ($ng_words->@*) {
        # ライブコメント一覧取得
        my $livecomments = $app->dbh->select_all_as(
            'Isupipe::Entity::Livecomment',
            'SELECT * FROM livecomments',
        );

        for my $livecomment ($livecomments->@*) {
            my $query = <<~ 'SQL';
                DELETE FROM livecomments
                WHERE
                id = ? AND
                livestream_id = ? AND
                (SELECT COUNT(*)
                FROM
                (SELECT ? AS text) AS texts
                INNER JOIN
                (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
                ON texts.text LIKE patterns.pattern) >= 1;
            SQL

            $app->dbh->query($query, $livecomment->id, $livestream_id, $livecomment->comment, $ng_word->word);
        }
    }

    $txn->commit;

    my $res = $c->render_json({ word_id => $word_id });
    $res->status(HTTP_CREATED);
    return $res;
}
