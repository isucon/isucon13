package Isupipe::Handler::TopHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::Tag;
use Isupipe::Entity::User;
use Isupipe::Entity::Theme;
use Isupipe::Util qw(
    verify_user_session
);

sub get_tag_handler($app, $c) {
    my $tags = $app->dbh->select_all_as(
        'Isupipe::Entity::Tag',
        'SELECT * FROM tags'
    );

    return $c->render_json({ tags => $tags });
}

# 配信者のテーマ取得API
sub get_streamer_theme_handler($app, $c) {
    verify_user_session($app, $c);

    my $username = $c->args->{username};

    my $txn = $app->dbh->txn_scope;

    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT id FROM users WHERE name = ?',
        $username
    );
    unless ($user) {
        $c->halt(HTTP_NOT_FOUND, 'user not found: '. $username);
    }

    my $theme = $app->dbh->select_row_as(
        'Isupipe::Entity::Theme',
        'SELECT * FROM themes WHERE user_id = ?',
        $user->id
    );

    $txn->commit;

    return $c->render_json($theme);
}

