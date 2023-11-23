use v5.38;
use experimental qw(class);

class Isupipe::Entity::LivecommentReport {
    field $id :param = undef;
    field $user_id :param = undef;
    field $livestream_id :param = undef;
    field $livecomment_id :param = undef;
    field $created_at :param = undef;
    field $reporter :param = undef;
    field $livecomment :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $user_id, 'user_id');
        assert_field(Int, $livestream_id, 'livestream_id');
        assert_field(Int, $livecomment_id, 'livecomment_id');
        assert_field(Int, $created_at, 'created_at');
        assert_field(InstanceOf['Isupipe::Entity::User'], $reporter, 'reporter');
        assert_field(InstanceOf['Isupipe::Entity::Livecomment'], $livecomment, 'livecomment');
    }

    method as_hashref() {
        return {
            id             => $id,
            user_id        => $user_id,
            livestream_id  => $livestream_id,
            livecomment_id => $livecomment_id,
            created_at     => $created_at,
        }
    }

    method TO_JSON() {
        return {
            id          => $id,
            reporter    => $reporter,
            livecomment => $livecomment,
            created_at  => $created_at,
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

    method livecomment_id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'livecomment_id'); $livecomment_id = $new };
        $livecomment_id
    }

    method created_at($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'created_at'); $created_at = $new };
        $created_at
    }

    method reporter($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::User'], $new, 'reporter'); $reporter = $new };
        $reporter
    }

    method livecomment($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::Livecomment'], $new, 'livecomment'); $livecomment = $new };
        $livecomment
    }
}
