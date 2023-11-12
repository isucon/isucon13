package Isupipe::App::ReactionHandler;
use v5.38;
use utf8;
use experimental qw(try);

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::Reaction;
use Isupipe::App::Util qw(
    verify_user_session
    DEFAULT_USER_ID_KEY
    check_params
);
use Isupipe::App::FillResponse qw(
    fill_reaction_response
);

use constant PostReactionRequest => Dict[
    emoji_name => Str,
];

sub get_reaction_handler($app, $c) {
    verify_user_session($app, $c);

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
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $params = $c->req->json_parameters;
    unless (check_params($params, PostReactionRequest)) {
        $c->halt_text(HTTP_BAD_REQUEST, 'invalid request');
    }

    my $txn = $app->dbh->txn_scope;
    try {
        my $reaction = Isupipe::Entity::Reaction->new(
            user_id       => $user_id,
            livestream_id => $livestream_id,
            emoji_name    => $params->{emoji_name},
            created_at    => time,
        );

        $app->dbh->query(
            'INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (:user_id, :livestream_id, :emoji_name,:created_at)',
            $reaction->as_hashref,
        );

        my $reaction_id = $app->dbh->last_insert_id;
        $reaction->id($reaction_id);

        $txn->commit;

        my $response = fill_reaction_response($app, $reaction);

        my $res = $c->render_json($response);
        $res->status(HTTP_CREATED);
        return $res;
    }
    catch($e) {
        $txn->rollback;
        if ($e isa Kossy::Exception) {
            die $e;
        }
        $c->halt_text(HTTP_INTERNAL_SERVER_ERROR, $e);
    }
}
