use v5.38;
use experimental qw(class);

class Isupipe::Entity::Reaction {
    field $id :param = undef;
    field $emoji_name :param = undef;
    field $user_id :param = undef;
    field $livestream_id :param = undef;
    field $created_at :param = undef;
    field $user :param = undef;
    field $livestream :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Str, $emoji_name, 'emoji_name');
        assert_field(Int, $user_id, 'user_id');
        assert_field(Int, $livestream_id, 'livestream_id');
        assert_field(Str, $created_at, 'created_at');
        assert_field(InstanceOf['Isupipe::Entity::User'], $user, 'user');
        assert_field(InstanceOf['Isupipe::Entity::Livestream'], $livestream, 'livestream');
    }

    method as_hashref() {
        return {
            id            => $id,
            emoji_name    => $emoji_name,
            user_id       => $user_id,
            livestream_id => $livestream_id,
            created_at    => $created_at,
        }
    }

    method TO_JSON() {
        return {
            id         => $id,
            emoji_name => $emoji_name,
            user       => $user,
            livestream => $livestream,
            created_at => $created_at,
        }
    }

    method id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'id'); $id = $new };
        $id
    }

    method emoji_name($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'emoji_name'); $emoji_name = $new };
        $emoji_name
    }

    method user_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'user_id'); $user_id = $new };
        $user_id
    }

    method livestream_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'livestream_id'); $livestream_id = $new };
        $livestream_id
    }

    method created_at($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'created_at'); $created_at = $new };
        $created_at
    }

    method user($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::User'], $new, 'user'); $user = $new };
        $user
    }

    method livestream($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::Livestream'], $new, 'livestream'); $livestream = $new };
        $livestream
    }
}
