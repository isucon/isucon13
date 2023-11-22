use v5.38;
use experimental qw(class);

class Isupipe::Entity::UserRankingEntry {
    field $user_name :param = undef;
    field $score :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Str, $user_name, 'user_name');
        assert_field(Int, $score, 'score');
    }

    method user_name($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'user_name'); $user_name = $new };
        $user_name
    }

    method score($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'score'); $score = $new };
        $score
    }
}
