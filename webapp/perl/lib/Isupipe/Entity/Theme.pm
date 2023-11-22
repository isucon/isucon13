use v5.38;
use experimental qw(class);

class Isupipe::Entity::Theme {
    field $id :param = undef;
    field $user_id :param = undef;
    field $dark_mode :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;
    use JSON::Types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $user_id, 'user_id');
        assert_field(Bool, $dark_mode, 'dark_mode');
    }

    method as_hashref() {
        return {
            id        => $id,
            user_id   => $user_id,
            dark_mode => $dark_mode,
        }
    }

    method TO_JSON() {
        return {
            id        => $id,
            dark_mode => bool $dark_mode,
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

    method dark_mode($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'dark_mode'); $dark_mode = $new };
        $dark_mode
    }
}
