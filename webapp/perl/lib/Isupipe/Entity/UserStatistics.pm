use v5.38;
use experimental qw(class);

class Isupipe::Entity::UserStatistics {
    field $rank :param = undef;
    field $viewers_count :param = undef;
    field $total_reactions :param = undef;
    field $total_livecomments :param = undef;
    field $total_tip :param = undef;
    field $favorite_emoji : param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $rank, 'rank');
        assert_field(Int, $viewers_count, 'viewers_count');
        assert_field(Int, $total_reactions, 'total_reactions');
        assert_field(Int, $total_livecomments, 'total_livecomments');
        assert_field(Int, $total_tip, 'total_tip');
        assert_field(Str, $favorite_emoji, 'favorite_emoji');
    }

    method as_hashref() {
        ...
    }

    method TO_JSON() {
        return {
            rank                => $rank,
            viewers_count       => $viewers_count,
            total_reactions     => $total_reactions,
            total_livecomments  => $total_livecomments,
            total_tip           => $total_tip,
            favorite_emoji      => $favorite_emoji,
        }
    }

    method rank($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'rank'); $rank = $new };
        $rank
    }

    method viewers_count($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'viewers_count'); $viewers_count = $new };
        $viewers_count
    }

    method total_reactions($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'total_reactions'); $total_reactions = $new };
        $total_reactions
    }

    method total_livecomments($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'total_livecomments'); $total_livecomments = $new };
        $total_livecomments
    }

    method total_tip($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'total_tip'); $total_tip = $new };
        $total_tip
    }

    method favorite_emoji($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'favorite_emoji'); $favorite_emoji = $new };
        $favorite_emoji
    }
}
