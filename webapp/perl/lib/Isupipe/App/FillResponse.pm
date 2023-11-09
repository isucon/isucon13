package Isupipe::App::FillResponse;
use v5.38;
use utf8;

use Exporter 'import';

our @EXPORT_OK = qw(
    fill_user_response
    fill_livestream_response
);

use Carp qw(croak);

use Isupipe::Entity::User;
use Isupipe::Entity::Tag;
use Isupipe::Entity::Livestream;
use Isupipe::Entity::LivestreamTag;

sub fill_user_response($app, $user) {
    my $theme = $app->dbh->select_row_as(
        'Isupipe::Entity::Theme',
        'SELECT * FROM themes WHERE user_id = ?',
        $user->id,
    );
    unless ($theme) {
        croak 'Theme not found:', $user->id;
    }

    return Isupipe::Entity::User->new(
        id          => $user->id,
        name        => $user->name,
        display_name => $user->display_name,
        description => $user->description,
        theme       => $theme,
    );
}

sub fill_livestream_response($app, $livestream) {
    my $owner = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE id = ?',
        $livestream->user_id,
    );
    unless ($owner) {
        croak 'Owner not found:', $livestream->user_id;
    }
    $owner = fill_user_response($app, $owner);

    my $livestream_tags = $app->dbh->select_all_as(
        'Isupipe::Entity::LivestreamTag',
        'SELECT * FROM livestream_tags WHERE livestream_id = ?',
        $livestream->id,
    );

    my $tags = [];
    for my $livestream_tag ($livestream_tags->@*) {
        my $tag = $app->dbh->select_row_as(
            'Isupipe::Entity::Tag',
            'SELECT * FROM tags WHERE id = ?',
            $livestream_tag->tag_id,
        );
        unless ($tag) {
            croak 'Tag not found:', $livestream_tag->tag_id;
        }
        push $tags->@*, $tag;
    }

    return Isupipe::Entity::Livestream->new(
        id            => $livestream->id,
        owner         => $owner,
        title         => $livestream->title,
        tags          => $tags,
        description   => $livestream->description,
        playlist_url  => $livestream->playlist_url,
        thumbnail_url => $livestream->thumbnail_url,
        start_at      => $livestream->start_at,
        end_at        => $livestream->end_at,
    );
}

