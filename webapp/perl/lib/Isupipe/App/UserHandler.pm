package Isupipe::App::UserHandler;
use v5.38;
use utf8;
use experimental qw(try);

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::User;
use Isupipe::Entity::Theme;
use Isupipe::App::Util qw(
    verify_user_session
    DEFAULT_USER_ID_KEY
    encrypt_password
    check_password
    check_params
);

use constant PostUserRequestTheme => Dict[
    dark_mode => Bool,
];

use constant PostUserRequest => Dict[
    name         => Str,
    display_name => Str,
    description  => Str,
    # Password is non-hashed password.
    password     => Str,
    theme        => PostUserRequestTheme,
];

use constant LoginRequest => Dict[
    name => Str,

    # Password is non-hashed password.
    password => Str,
];

# ユーザ登録API
# POST /api/register
sub register_handler($app, $c) {
    my $params = $c->req->json_parameters;
    unless (check_params($params, PostUserRequest)) {
        $c->halt(HTTP_BAD_REQUEST, 'failed to decode the quest body as json');
    }

    if ($params->{name} eq 'pipe') {
        $c->halt(HTTP_BAD_REQUEST, "the username 'pipe' is reserved");
    }

    my $hashed_password = encrypt_password($params->{password});

    my $txn = $app->dbh->txn_scope;
    try {
        my $user = Isupipe::Entity::User->new(
            name         => $params->{name},
            display_name => $params->{display_name},
            description  => $params->{description},
            password     => $hashed_password,
        );

        $app->dbh->query(
            'INSERT INTO users (name, display_name, description, password) VALUES(:name, :display_name, :description, :password)',
            $user->as_hashref
        );

        my $user_id = $app->dbh->last_insert_id;
        $user->id($user_id);

        my $theme = Isupipe::Entity::Theme->new(
            user_id   => $user_id,
            dark_mode => $params->{theme}{dark_mode},
        );

        $app->dbh->query(
            'INSERT INTO themes (user_id, dark_mode) VALUES(:user_id, :dark_mode)',
            $theme->as_hashref
        );

        $txn->commit;

        my $res = $c->render_json($user);
        $res->status(HTTP_CREATED);
        return $res;
    }
    catch ($e) {
        $txn->rollback;
        if ($e isa Kossy::Exception) {
            die $e;
        }
        $c->halt(HTTP_INTERNAL_SERVER_ERROR, $e);
    }
}


# ユーザログインAPI
# POST /api/login
sub login_handler($app, $c) {
    my $params = $c->req->json_parameters;
    unless (check_params($params, LoginRequest)) {
        $c->halt_text(HTTP_BAD_REQUEST, 'failed to decode the quest body as json');
    }

    my $txn = $app->dbh->txn_scope;

    # usernameはUNIQUEなので、whereで一意に特定できる
    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = :name',
        { name => $params->{name} }
    );
    unless ($user) {
        $c->halt_text(HTTP_NOT_FOUND, 'invalid username or password');
    }

    $txn->commit;

    unless (check_password($params->{password}, $user->password)) {
        $c->halt_text(HTTP_UNAUTHORIZED, 'invalid username or password');
    }

    my $session = Plack::Session->new($c->env);
    $session->set(DEFAULT_USER_ID_KEY, $user->id);

    return $c->halt_no_content(HTTP_OK);
}


# ユーザ詳細API
# GET /user/:user_id
sub user_handler($app, $c) {
    verify_user_session($app, $c);

    my $user_id = $c->args->{user_id};

    my $user = $app->db->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE id = :id',
        { id => $user_id }
    );
    unless ($user) {
        $c->halt(HTTP_NOT_FOUND, 'user not found');
    }

    return $c->render_json($user);
}

# 配信者のテーマ取得API
# GET /user/:userid/theme
sub get_user_theme_handler($app, $c) {
    verify_user_session($app, $c);

    my $user_id = $c->args->{user_id};
    my $theme = $app->db->select_row_as(
        'Isupipe::Entity::Theme',
        'SELECT * FROM themes WHERE user_id = ?',
        $user_id
    );
    unless ($theme) {
        $c->halt(HTTP_NOT_FOUND, 'theme not found');
    }

    return $c->render_json($theme);
}

sub get_users_handler($app, $c) {
    my $users = $app->db->select_all_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users'
    );

    return $c->render_json($users);
}


