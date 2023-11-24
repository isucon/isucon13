package Isupipe::Handler::UserHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;
use Plack::Session;;
use MIME::Base64 qw(decode_base64);

use Isupipe::Log;
use Isupipe::Entity::User;
use Isupipe::Entity::Theme;
use Isupipe::Util qw(
    verify_user_session
    DEFAULT_SESSION_EXPIRES_KEY
    DEFAULT_USER_ID_KEY
    DEFAULT_USER_NAME_KEY
    encrypt_password
    check_password
    check_params
);
use Isupipe::FillResponse qw(
    fill_user_response
);

use Isupipe::Icon qw(
    read_fallback_user_icon_image
);

use constant POWER_DNS_SUBDMAIN_ADDRESS => $ENV{ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS};

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
    username => Str,

    # Password is non-hashed password.
    password => Str,
];

use constant PostIconRequest => Dict[
    image => Str, # 画像データをBase64した値
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


    my $err = system("pdnsutil", "add-record", "u.isucon.dev", $params->{name}, "A", "0", POWER_DNS_SUBDMAIN_ADDRESS);
    if ($err) {
        $c->halt(HTTP_INTERNAL_SERVER_ERROR, $err);
    }

    $user = fill_user_response($app, $user);

    $txn->commit;

    my $res = $c->render_json($user);
    $res->status(HTTP_CREATED);
    return $res;
}


# ユーザログインAPI
# POST /api/login
sub login_handler($app, $c) {
    my $params = $c->req->json_parameters;
    unless (check_params($params, LoginRequest)) {
        $c->halt(HTTP_BAD_REQUEST, 'failed to decode the quest body as json');
    }

    my $txn = $app->dbh->txn_scope;

    # usernameはUNIQUEなので、whereで一意に特定できる
    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = :name',
        { name => $params->{username} }
    );
    unless ($user) {
        $c->halt(HTTP_UNAUTHORIZED, 'invalid username or password');
    }

    $txn->commit;

    unless (check_password($params->{password}, $user->password)) {
        $c->halt(HTTP_UNAUTHORIZED, 'invalid username or password');
    }

    my $session = Plack::Session->new($c->env);

    $session->set(DEFAULT_USER_ID_KEY, $user->id);
    $session->set(DEFAULT_USER_NAME_KEY, $user->name);
    $session->set(DEFAULT_SESSION_EXPIRES_KEY, time + 3600);

    $c->halt_no_content(HTTP_OK);
}

sub get_me_handler($app, $c) {
    verify_user_session($app, $c);

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $txn = $app->dbh->txn_scope;

    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE id = ?',
        $user_id,
    );
    unless ($user) {
        $c->halt(HTTP_NOT_FOUND, 'not found user that has the userid in session');
    }

    $user = fill_user_response($app, $user);

    $txn->commit;

    my $res = $c->render_json($user);
    return $res;
}

# ユーザー詳細API
# GET /api/user/:username
sub get_user_handler($app, $c) {
    verify_user_session($app, $c);

    my $username = $c->args->{username};

    my $txn = $app->dbh->txn_scope;

    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = ?',
        $username,
    );
    unless ($user) {
        $c->halt(HTTP_NOT_FOUND, 'not found user that has the given username');
    }

    $user = fill_user_response($app, $user);

    $txn->commit;

    return $c->render_json($user);
}


sub get_icon_handler($app, $c) {
    my $username = $c->args->{username};

    my $txn = $app->dbh->txn_scope;
    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = ?',
        $username,
    );
    unless ($user) {
        $c->halt(HTTP_NOT_FOUND, 'not found user that has the given username');
    }

    my $image = $app->dbh->select_one(
        'SELECT image FROM icons WHERE user_id = ?',
        $user->id,
    );

    if (!$image) {
        $image = read_fallback_user_icon_image();
    }

    $txn->commit;

    my $res = $c->response;
    $res->status(HTTP_OK);
    $res->content_type('image/jpeg');
    $res->body($image);
    return $res;
}

sub post_icon_handler($app, $c) {
    verify_user_session($app, $c);

    # existence already checked
    my $user_id = $c->req->session->{+DEFAULT_USER_ID_KEY};

    my $params = $c->req->json_parameters;
    unless (check_params($params, PostIconRequest)) {
        $c->halt(HTTP_BAD_REQUEST, 'failed to decode the quest body as json');
    }

    my $txn = $app->dbh->txn_scope;
    $app->dbh->query(
        'DELETE FROM icons WHERE user_id = ?', $user_id
    );

    $app->dbh->query(
        'INSERT INTO icons (user_id, image) VALUES(?, ?)',
        $user_id,
        decode_base64($params->{image}),
    );

    my $icon_id = $app->dbh->last_insert_id;

    $txn->commit;

    my $res = $c->render_json({ id => $icon_id });
    $res->status(HTTP_CREATED);
    return $res;
}
