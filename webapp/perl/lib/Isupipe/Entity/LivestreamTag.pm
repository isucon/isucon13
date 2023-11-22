use v5.38;
use experimental qw(class);

class Isupipe::Entity::LivestreamTag {
    field $id :param = undef;
    field $livestream_id :param = undef;
    field $tag_id :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $livestream_id, 'livestream_id');
        assert_field(Int, $tag_id, 'tag_id');
    }

    method as_hashref() {
        return {
            id            => $id,
            livestream_id => $livestream_id,
            tag_id        => $tag_id,
        }
    }

    method TO_JSON() {
        ...
    }

    method id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'id'); $id = $new };
        $id
    }

    method livestream_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'livestream_id'); $livestream_id = $new };
        $livestream_id
    }

    method tag_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'tag_id'); $tag_id = $new };
        $tag_id
    }
}
