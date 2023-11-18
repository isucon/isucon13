use v5.38;
use experimental qw(class);

class Isupipe::Entity::Livecomment {
    field $id :param = undef;
    field $user_id :param = undef;
    field $user :param = undef;
    field $livestream_id :param = undef;
    field $livestream :param = undef;
    field $comment :param = undef;
    field $tip :param = undef;
    field $created_at :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;
    use JSON::Types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $user_id, 'user_id');
        assert_field(InstanceOf['Isupipe::Entity::User'], $user, 'user');
        assert_field(Int, $livestream_id, 'livestream_id');
        assert_field(InstanceOf['Isupipe::Entity::Livestream'], $livestream, 'livestream');
        assert_field(Str, $comment, 'comment');
        assert_field(Int, $tip, 'tip');
        assert_field(Str, $created_at, 'created_at');
    }

    method as_hashref() {
        return {
            id            => $id,
            user_id       => $user_id,
            livestream_id => $livestream_id,
            comment       => $comment,
            tip           => $tip,
            created_at    => $created_at,
        }
    }

    method TO_JSON() {
        return {
            id            => $id,
            user          => $user,
            livestream    => $livestream,
            comment       => $comment,
            tip           => number $tip,
            created_at    => $created_at,
        };
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

    method livestream($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::Livestream'], $new, 'livestream'); $livestream = $new };
        $livestream
    }

    method comment($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'comment'); $comment = $new };
        $comment
    }

    method tip($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'tip'); $tip = $new };
        $tip
    }

    method created_at($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'created_at'); $created_at = $new };
        $created_at
    }
}
