package Isupipe::App::StatsHandler;
use v5.38;
use utf8;

# FIXME: 配信毎、ユーザごとのリアクション種別ごとの数などもだす
use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::LivestreamStatistics;
use Isupipe::Entity::UserStatistics;
use Isupipe::Entity::LivestreamViewer;
use Isupipe::App::Util qw(
    verify_user_session
    check_params
);

sub get_user_statistics_handler($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $user_id = $c->args->{user_id};

    my $viewed_livestreams = $app->dbh->select_all_as(
        'Isupipe::Entity::LivestreamViewer',
        'SELECT user_id, livestream_id FROM livestream_viewers_history WHERE user_id = ?',
        $user_id
    );

    my $tip_rank_per_livestreams = query_total_tip_rank_per_viewed_livestream($app, $c, $user_id, $viewed_livestreams);

    my $stats = Isupipe::Entity::UserStatistics->new(
        tip_rank_per_livestreams => $tip_rank_per_livestreams,
    );

    return $c->render_json($stats);
}

sub get_livestream_statistics_handler($app, $c) {
    my $err = verify_user_session($app, $c);
    if ($err isa Kossy::Exception) {
        die $err;
    }

    my $livestream_id = $c->args->{livestream_id};

    my $query = <<~QUERY;
        SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank
        FROM livecomments GROUP BY livestream_id
        HAVING livestream_id = ? ORDER BY total_tip DESC LIMIT 3
    QUERY

    my $tip_ranks = $app->dbh->select_all_as(
        'Isupipe::Entity::TipRank',
        $query,
        $livestream_id
    );

    $query = <<~QUERY;
        SELECT COUNT(*) AS total_reaction, emoji_name, RANK() OVER(ORDER BY COUNT(*) DESC) AS reaction_rank
        FROM reactions GROUP BY livestream_id, emoji_name
        HAVING livestream_id = ? ORDER BY total_reaction DESC LIMIT 3
    QUERY

    my $reaction_ranks = $app->dbh->select_all_as(
        'Isupipe::Entity::ReactionRank',
        $query,
        $livestream_id
    );

    my $stats = Isupipe::Entity::LivestreamStatistics->new(
        most_tip_ranking => $tip_ranks,
        most_posted_reaction_ranking => $reaction_ranks,
    );

    return $c->render_json($stats);
}


sub query_total_top_rank_perl_viewed_livestream($app, $c, $user_id, $viewed_livestreams) {

    my $total_tip_rank_per_livestreams = {};

    # get total tip per viewed livestream
    for my $viewed_livestream ($viewed_livestreams->@*) {
        my $total_tip = $app->dbh->select_one(
            'SELECT SUM(tip) FROM livecomments WHERE user_id = ? AND livestream_id = ?',
            $viewed_livestream->user_id,
            $viewed_livestream->livestream_id
        );

        my $total_tip_rank_perl_livestream->{$viewed_livestream->livestream_id} = Isupipe::Entity::TipRank->new(
            total_tip => $total_tip,
        );
    }

    for my $livestream_id (keys $total_tip_rank_per_livestreams->%*) {
        my $stat = $total_tip_rank_per_livestreams->{$livestream_id};

        my $query = <<~QUERY;
            SELECT SUM(tip) AS total_tip, RANK() OVER(ORDER BY SUM(tip) DESC) AS tip_rank
            FROM livecomments GROUP BY livestream_id
            HAVING livestream_id = ? AND total_tip = ?
        QUERY

        my $rank = $app->dbh->select_one_as(
            'Isupipe::Entity::TipRank',
            $query,
            $livestream_id,
            $stat->total_tip
        );

        $total_tip_rank_per_livestreams->{$livestream_id} = $rank;
    }

    return $total_tip_rank_per_livestreams;
}
