package Isupipe::App::ReactionHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Entity::Reaction;

use Isupipe::App::UserHandler qw(
    verify_user_session
    DEFAULT_USER_ID_KEY
);

use constant PostReactionRequest => Dict[
    emoji_name => Str,
];

sub get_reaction_handler($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $livestream_id = $c->args->{livestream_id};

    my $reactions = $app->dbh->select_all_as(
        'Isupipe::Entity::Reaction',
        'SELECT * FROM reactions WHERE livestream_id = ?',
        $livestream_id,
    );
    unless (@$reactions) {
        $c->halt_text(HTTP_NOT_FOUND, 'reactions not found');
    }

    return $c->render_json($reactions);
}

sub post_reaction_handler($app, $c) {
    my $livestream_id = $c->args->{livestream_id};

    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $user_id = $c->req->session->get(DEFAULT_USER_ID_KEY);
    unless ($user_id) {
        $c->halt_text(HTTP_UNAUTHORIZED, 'failed to find user-id from session');
    }

    my $req = $c->req->body_parameters;
    unless (PostReactionRequest->check($req)) {
        # FIXME: GO言語の参考実装では、500を返している
        $c->halt_text(HTTP_BAD_REQUEST, 'invalid request');
    }

    my $dbh = $app->dbh;
    my $txn = $dbh->txn_scope;

    my $reaction = Isupipe::Entity::Reaction->new(
        user_id       => $user_id,
        livestream_id => $livestream_id,
        emoji_name    => $req->{emoji_name},
    );

    $dbh->query(
        'INSERT INTO reactions (user_id, livestream_id, emoji_name) VALUES (:user_id, :livestream_id, :emoji_name)',
        $reaction->as_hashref,
    );

    my $reaction_id = $dbh->last_insert_id;

    $txn->commit;

    $reaction->set_id($reaction_id);

    my $res = $c->render_json($reaction);
    $res->status(HTTP_CREATED);
    return $res;
}
