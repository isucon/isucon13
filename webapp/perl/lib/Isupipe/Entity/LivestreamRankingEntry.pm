use v5.38;
use experimental qw(class);

class Isupipe::Entity::LivestreamRankingEntry {
    field $livestream_id :param = undef;
    field $score :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $livestream_id, 'livestream_id');
        assert_field(Int, $score, 'score');
    }

    method livestream_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'livestream_id'); $livestream_id = $new };
        $livestream_id
    }

    method score($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'score'); $score = $new };
        $score
    }
}
