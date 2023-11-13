package Isupipe::Handler::ReactionHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::Reaction;
use Isupipe::Util qw(
    verify_user_session
    DEFAULT_USER_ID_KEY
    check_params
);
use Isupipe::FillResponse qw(
    fill_reaction_response
);

use constant PostReactionRequest => Dict[
    emoji_name => Str,
];

sub get_reactions_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    my $txn = $app->dbh->txn_scope;

    my $query = 'SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC';
    if (my $limit = $c->req->query_parameters->{limit}) {
        unless ($limit =~ /^\d+$/) {
            $c->halt(HTTP_BAD_REQUEST, "limit query parameter must be integer");
        }
        $query .= sprintf(" LIMIT %d", $limit);
    }

    my $reactions = $app->dbh->select_all_as(
        'Isupipe::Entity::Reaction',
        $query,
        $livestream_id,
    );

    $reactions = [
        map { fill_reaction_response($app, $_) } @$reactions
    ];

    $txn->commit;

    return $c->render_json($reactions);
}

sub post_reaction_handler($app, $c) {
    verify_user_session($app, $c);

    my $livestream_id = $c->args->{livestream_id};

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $params = $c->req->json_parameters;
    unless (check_params($params, PostReactionRequest)) {
        $c->halt(HTTP_BAD_REQUEST, 'invalid request');
    }

    my $txn = $app->dbh->txn_scope;

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

    $reaction = fill_reaction_response($app, $reaction);

    my $res = $c->render_json($reaction);
    $res->status(HTTP_CREATED);
    return $res;
}
