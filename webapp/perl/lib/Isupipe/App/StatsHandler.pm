package Isupipe::App::StatsHandler;
use v5.38;
use utf8;

# FIXME: 配信毎、ユーザごとのリアクション種別ごとの数などもだす
use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::UserRankingEntry;
use Isupipe::Entity::UserStatistics;
use Isupipe::Entity::LivestreamStatistics;

use Isupipe::App::Util qw(
    verify_user_session
    check_params
);

sub get_user_statistics_handler($app, $c) {
    verify_user_session($app, $c);

    my $username = $c->args->{username};

    my $txn = $app->dbh->txn_scope;

    my $selected_user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = ?',
        $username,
    );
    unless ($selected_user) {
        $c->halt(HTTP_NOT_FOUND, 'not found user that has the given username');
    }

    # ランク算出
    my $users = $app->dbh->select_all_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users'
    );

    my $ranking = [];
    for my $user ($users->@*) {

        my $reactions = $app->dbh->select_one(
            q[
                SELECT COUNT(*) FROM users u
                INNER JOIN livestreams l ON l.user_id = u.id
                INNER JOIN reactions r ON r.livestream_id = l.id
                WHERE u.id = ?
            ],
            $user->id
        );

        my $tips = $app->dbh->select_one(
            q[
                SELECT IFNULL(SUM(l2.tip), 0) FROM users u
                INNER JOIN livestreams l ON l.user_id = u.id
                INNER JOIN livecomments l2 ON l2.livestream_id = l.id
                WHERE u.id = ?
            ],
            $user->id
        );

        my $score = $reactions + $tips;
        push $ranking->@* => Isupipe::Entity::UserRankingEntry->new(
            user_name => $user->name,
            score => $score,
        );
    }

    my @sorted_ranking = sort {
        if ($a->score == $b->score) {
            $a->user_name cmp $b->user_name;
        }
        else {
            $a->score <=> $b->score;
        }
    } $ranking->@*;

    $ranking = \@sorted_ranking;

    my $rank = 1;
    for (my $i = scalar $ranking->@* - 1; $i >= 0; $i--) {
        my $entry = $ranking->[$i];
        if ($entry->user_name eq $username) {
            last;
        }
        $rank++;
    }

    # リアクション数
    my $total_reactions = $app->dbh->select_one(
        q[
            SELECT COUNT(*) FROM users u 
            INNER JOIN livestreams l ON l.user_id = u.id 
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
        ],
        $username,
    );

    # ライブコメント数、チップ合計
    my $total_livecomments = 0;
    my $total_tip = 0;
    for my $user ($users->@*) {
        my $livestreams = $app->dbh->select_all_as(
            'Isupipe::Entity::Livestream',
            'SELECT * FROM livestreams WHERE user_id = ?',
            $user->id,
        );

        for my $livestream ($livestreams->@*) {
            my $livecomments = $app->dbh->select_all_as(
                'Isupipe::Entity::Livecomment',
                'SELECT * FROM livecomments WHERE livestream_id = ?',
                $livestream->id,
            );

            for my $livecomment ($livecomments->@*) {
                $total_tip += $livecomment->tip;
                $total_livecomments++;
            }
        }
    }

    # 合計視聴者数
    my $viewers_count = 0;
    for my $user ($users->@*) {
        my $livestreams = $app->dbh->select_all_as(
            'Isupipe::Entity::Livestream',
            'SELECT * FROM livestreams WHERE user_id = ?',
            $user->id,
        );

        for my $livestream ($livestreams->@*) {
            my $cnt = $app->dbh->select_one(
                'SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?',
                $livestream->id,
            );
            $viewers_count += $cnt;
        }
    }

    # お気に入り絵文字
    my $favorite_emoji = $app->dbh->select_one(
        q[
            SELECT r.emoji_name
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
            GROUP BY emoji_name
            ORDER BY COUNT(*) DESC
            LIMIT 1
        ],
        $username
    );

    my $stats = Isupipe::Entity::UserStatistics->new(
        rank               => $rank,
        viewers_count      => $viewers_count,
        total_reactions    => $total_reactions,
        total_livecomments => $total_livecomments,
        total_tip          => $total_tip,
        favorite_emoji     => $favorite_emoji,
    );

    return $c->render_json($stats);
}

sub get_livestream_statistics_handler($app, $c) {
    verify_user_session($app, $c);

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
