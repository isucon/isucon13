use v5.38;
use experimental qw(class);

class Isupipe::Entity::LivestreamStatistics {
    field $rank :param = undef;
    field $viewers_count :param = undef;
    field $total_reactions :param = undef;
    field $total_reports :param = undef;
    field $max_tip :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $rank, 'rank');
        assert_field(Int, $viewers_count, 'viewers_count');
        assert_field(Int, $total_reactions, 'total_reactions');
        assert_field(Int, $total_reports, 'total_reports');
        assert_field(Int, $max_tip, 'max_tip');
    }

    method as_hashref() {
        ...
    }

    method TO_JSON() {
        return {
            rank            => $rank,
            viewers_count   => $viewers_count,
            total_reactions => $total_reactions,
            total_reports   => $total_reports,
            max_tip         => $max_tip,
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

    method total_reports($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'total_reports'); $total_reports = $new };
        $total_reports
    }

    method max_tip($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'max_tip'); $max_tip = $new };
        $max_tip
    }
}
