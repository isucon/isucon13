use v5.38;
use experimental qw(class);

class Isupipe::Entity::NGWord {
    field $id :param = undef;
    field $user_id :param = undef;
    field $livestream_id :param = undef;
    field $word :param = undef;
    field $created_at :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $user_id, 'user_id');
        assert_field(Int, $livestream_id, 'livestream_id');
        assert_field(Str, $word, 'word');
        assert_field(Int, $created_at, 'created_at');
    }

    method as_hashref() {
        return {
            id            => $id,
            user_id       => $user_id,
            livestream_id => $livestream_id,
            word          => $word,
            created_at    => $created_at,
        }
    }

    method TO_JSON() {
        return {
            id            => $id,
            user_id       => $user_id,
            livestream_id => $livestream_id,
            word          => $word,
            created_at    => $created_at,
        }
    }

    method id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'id'); $id = $new };
        $id
    }

    method user_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'user_id'); $user_id = $new };
        $user_id
    }

    method livestream_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'livestream_id'); $livestream_id = $new };
        $livestream_id
    }

    method word($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'word'); $word = $new };
        $word
    }

    method created_at($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'created_at'); $created_at = $new };
        $created_at
    }
}
