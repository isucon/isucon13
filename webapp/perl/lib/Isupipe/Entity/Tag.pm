use v5.38;
use experimental qw(class);

class Isupipe::Entity::Tag {
    field $id :param = undef;
    field $name :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Str, $name, 'name');
    }

    method as_hashref() {
        return {
            id   => $id,
            name => $name,
        }
    }

    method TO_JSON() {
        return {
            id   => $id,
            name => $name,
        };
    }

    method id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'id'); $id = $new };
        $id
    }

    method name($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'name'); $name = $new };
        $name
    }
}
