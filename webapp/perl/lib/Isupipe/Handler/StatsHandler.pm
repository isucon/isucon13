package Isupipe::Handler::StatsHandler;
use v5.38;
use utf8;

use HTTP::Status qw(:constants);
use Types::Standard -types;

use Isupipe::Log;
use Isupipe::Entity::UserRankingEntry;
use Isupipe::Entity::UserStatistics;
use Isupipe::Entity::LivestreamRankingEntry;
use Isupipe::Entity::LivestreamStatistics;

use Isupipe::Util qw(
    verify_user_session
    check_params
);

sub get_user_statistics_handler($app, $c) {
    verify_user_session($app, $c);

    my $username = $c->args->{username};
    # ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
    # また、現在の合計視聴者数もだす

    my $txn = $app->dbh->txn_scope;

    my $user = $app->dbh->select_row_as(
        'Isupipe::Entity::User',
        'SELECT * FROM users WHERE name = ?',
        $username,
    );
    unless ($user) {
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

    # 合計視聴者数
    my $viewers_count = 0;

    for my $livestream ($livestreams->@*) {
        my $cnt = $app->dbh->select_one(
            'SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?',
            $livestream->id,
        );
        $viewers_count += $cnt;
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
            ORDER BY COUNT(*) DESC, emoji_name DESC
            LIMIT 1
        ],
        $username
    );

    $txn->commit;

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

    my $txn = $app->dbh->txn_scope;

    my $selected_livestream = $app->dbh->select_row_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams WHERE id = ?',
        $livestream_id,
    );
    unless ($selected_livestream) {
        $c->halt(HTTP_NOT_FOUND, 'cannot get stats of not found livestream');
    }

    my $livestreams = $app->dbh->select_all_as(
        'Isupipe::Entity::Livestream',
        'SELECT * FROM livestreams'
    );

    # ランク算出
    my $ranking = [];
    for my $livestream ($livestreams->@*) {
        my $reactions = $app->dbh->select_one(
            q[
                SELECT COUNT(*) FROM livestreams l
                INNER JOIN reactions r ON r.livestream_id = l.id
                WHERE l.id = ?
            ],
            $livestream->id
        );

        my $total_tips = $app->dbh->select_one(
            q[
                SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l
                INNER JOIN livecomments l2 ON l2.livestream_id = l.id
                WHERE l.id = ?
            ],
            $livestream->id
        );

        my $score = $reactions + $total_tips;
        push $ranking->@* => Isupipe::Entity::LivestreamRankingEntry->new(
            livestream_id => $livestream->id,
            score => $score,
        );
    }

    my @sorted_ranking = sort {
        if ($a->score == $b->score) {
            $a->livestream_id <=> $b->livestream_id;
        }
        else {
            $a->score <=> $b->score;
        }
    } $ranking->@*;

    $ranking = \@sorted_ranking;

    my $rank = 1;
    for (my $i = scalar $ranking->@* - 1; $i >= 0; $i--) {
        my $entry = $ranking->[$i];
        if ($entry->livestream_id == $livestream_id) {
            last;
        }
        $rank++;
    }

    # 視聴者数算出
    my $viewers_count = $app->dbh->select_one(
        q[SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?],
        $livestream_id,
    );

    # 最大チップ額
    my $max_tip = $app->dbh->select_one(
        q[SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?],
        $livestream_id,
    );
    # リアクション数
    my $total_reactions = $app->dbh->select_one(
        q[SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?],
        $livestream_id,
    );

    # スパム報告数
    my $total_reports = $app->dbh->select_one(
        q[SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?],
        $livestream_id,
    );

    $txn->commit;

    my $stats = Isupipe::Entity::LivestreamStatistics->new(
        rank            => $rank,
        viewers_count   => $viewers_count,
        max_tip         => $max_tip,
        total_reactions => $total_reactions,
        total_reports   => $total_reports,
    );

    return $c->render_json($stats);
}

