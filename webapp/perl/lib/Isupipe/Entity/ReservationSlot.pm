use v5.38;
use experimental qw(class);

class Isupipe::Entity::ReservationSlot {
    field $id :param = undef;
    field $slot :param = undef;
    field $start_at :param = undef;
    field $end_at :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $slot, 'slot');
        assert_field(Int, $start_at, 'start_at');
        assert_field(Int, $end_at, 'end_at');
    }

    method as_hashref() {
        return {
            id       => $id,
            slot     => $slot,
            start_at => $start_at,
            end_at   => $end_at,
        }
    }

    method TO_JSON() {
        ...
    }

    method id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'id'); $id = $new };
        $id
    }

    method slot($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'slot'); $slot = $new };
        $slot
    }

    method start_at($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'start_at'); $start_at = $new };
        $start_at
    }

    method end_at($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'end_at'); $end_at = $new };
        $end_at
    }
}
