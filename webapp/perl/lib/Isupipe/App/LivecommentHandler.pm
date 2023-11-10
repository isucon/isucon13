package Isupipe::App::LivecommentHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::Livecomment;
use Isupipe::Entity::LivecommentReport;
use Isupipe::Entity::NGWord;
use Isupipe::App::Util qw(
    verify_user_session
    DEFAULT_USER_ID_KEY
    check_params
);

use constant PostLivecommentRequest => Dict[
    comment => Str,
    tip => Int,
];

use constant ModerateRequest => Dict[
    ng_word => Str,
];

sub get_livecomment_handler($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $livestream_id = $c->args->{livestream_id};
    my $livecomments = $app->dbh->select_all_as(
        'Isupipe::Entity::Livecomment',
        'SELECT * FROM livecomments WHERE livestream_id = ?',
        $livestream_id
    );

    unless (@$livecomments) {
        $c->halt_text(HTTP_NOT_FOUND, "livestream not found");
    }

    return $c->render_json($livecomments);
}

sub post_livetcomment_handleer($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $livestream_id = $c->args->{livestream_id};

    my $user_id = $c->req->session->{DEFAULT_USER_ID_KEY};
    unless ($user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, "failed to find user-id from session");
    }

    my $params = $c->req->json_parameters;
    unless (check_params($params, PostLivecommentRequest)) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad request");
    }

    if ($params->{tip} < 0 || 20000 < $params->{tip}) {
        # FIXME コメントと式が食い違ってる 1 <= tips < 20000 が正しい？
        $c->halt_text(HTTP_BAD_REQUEST, "the tips in a live comment be 1 <= tips <= 20000");
    }

    my $dbh = $app->dbh;
    my $txn = $dbh->txn_scope;

    my $query =<<~'QUERY';
    SELECT COUNT(*) AS cnt
    FROM ng_words AS w
    CROSS JOIN
    (SELECT ? AS text) AS t
    WHERE t.text LIKE CONCAT('%', w.word, '%');
    QUERY

    my $hit_spam = $dbh->select_one($query, $params->{comment});
    infof("[hitSpam=%d] comment=%s", $hit_spam, $params->{comment});
    if ($hit_spam >= 1) {
        $c->halt_text(HTTP_BAD_REQUEST, "このコメントがスパム判定されました");
    }

    my $livecomment = Isupipe::Entity::Livecomment->new(
        user_id => $user_id,
        livestream_id => $livestream_id,
        comment => $params->{comment},
        tip => $params->{tip},
    );

    $dbh->query(
        'INSERT INTO livecomments (user_id, livestream_id, comment, tip) VALUES (:user_id, :livestream_id, :comment, :tip)',
        $livecomment->as_hashref,
    );

    my $livecomment_id = $dbh->last_insert_id;

    $txn->commit;

    $livecomment->set_id($livecomment_id);
    my $created_at = time;
    $livecomment->set_created_at($created_at);
    $livecomment->set_updated_at($created_at);

    return $c->render_json($livecomment);
}

sub report_livecomment_handler($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $livestream_id = $c->args->{livestream_id};
    my $livecomment_id = $c->args->{livecomment_id};

    my $user_id = $c->req->session->{DEFAULT_USER_ID_KEY};
    unless ($user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, "failed to find user-id from session");
    }

    my $dbh = $app->dbh;
    my $txn = $dbh->txn_scope;

    # 配信者自身の配信に対するGETなのかを検証
    my $owned_livestreams = $dbh->select_all(
        'SELECT * FROM livestreams WHERE id = ? AND user_id = ?',
        $livestream_id,
        $user_id,
    );
    if (@$owned_livestreams == 0) {
        $c->halt_text(HTTP_BAD_REQUEST, "A streamer can't get livecomment reports that other streamers own")
    }

    $dbh->query(
        'INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id) VALUES (:user_id, :livestream_id, :livecomment_id)',
        {
            user_id => $user_id,
            livestream_id => $livestream_id,
            livecomment_id => $livecomment_id,
        },
    );

    $dbh->query(
        'UPDATE livecomments SET report_count = report_count + 1 WHERE id = ?',
        $livecomment_id,
    );

    my $report_id = $dbh->last_insert_id;

    my $created_at = time;
    my $report = Isupipe::Entity::LivecommentReport->new(
        id => $report_id,
        user_id => $user_id,
        livestream_id => $livestream_id,
        livecomment_id => $livecomment_id,
        created_at => $created_at,
        updated_at => $created_at,
    );

    $txn->commit;

    my $res = $c->render_json($report);
    $res->status(HTTP_CREATED);
    return $res;
}

# NGワードを登録
sub moderate_ng_word_handler($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $livestream_id = $c->args->{livestream_id};
    my $user_id = $c->req->session->{DEFAULT_USER_ID_KEY};
    unless ($user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, "failed to find user-id from session");
    }

    my $params = $c->req->json_parameters;
    unless (check_params($params, ModerateRequest)) {
        $c->halt_text(HTTP_BAD_REQUEST, "bad request");
    }

    my $dbh = $app->dbh;
    my $txn = $dbh->txn_scope;

    # 配信者自身の廃止に対するmoderateなのかを検証
    my $owned_livestreams = $dbh->select_all(
        'SELECT * FROM livestreams WHERE id = ? AND user_id = ?',
        $livestream_id,
        $user_id,
    );
    if (@$owned_livestreams == 0) {
        $c->halt_text(HTTP_BAD_REQUEST, "A streamer can't get livecomment reports that other streamers own")
    }

    $dbh->query(
        'INSERT INTO ng_words(user_id, livestream_id, word) VALUES (:user_id, :livestream_id, :word)',
        {
            user_id => $user_id,
            livestream_id => $livestream_id,
            word => $params->{ng_word},
        },
    );

    my $word_id = $dbh->last_insert_id;
    $txn->commit;

    my $res = $c->render_json({ word_id => $word_id });
    $res->status(HTTP_CREATED);
    return $res;
}
