use v5.38;
use experimental qw(class);

class Isupipe::Entity::Livestream {
    field $id :param = undef;
    field $user_id :param = undef;
    field $title :param = undef;
    field $description :param = undef;
    field $playlist_url :param = undef;
    field $thumbnail_url :param = undef;
    field $start_at :param = undef;
    field $end_at :param = undef;

    field $owner :param = undef;
    field $tags :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Int, $user_id, 'user_id');
        assert_field(Str, $title, 'title');
        assert_field(Str, $description, 'description');
        assert_field(Str, $playlist_url, 'playlist_url');
        assert_field(Str, $thumbnail_url, 'thumbnail_url');
        assert_field(Int, $start_at, 'start_at');
        assert_field(Int, $end_at, 'end_at');

        assert_field(InstanceOf['Isupipe::Entity::User'], $owner, 'owner');
        assert_field(ArrayRef[InstanceOf['Isupipe::Entity::Tag']], $tags, 'tags');
    }

    method as_hashref() {
        return {
            id            => $id,
            user_id       => $user_id,
            title         => $title,
            playlist_url  => $playlist_url,
            thumbnail_url => $thumbnail_url,
            description   => $description,
            start_at      => $start_at,
            end_at        => $end_at,
        }
    }

    method TO_JSON() {
        return {
            id            => $id,
            owner         => $owner,
            title         => $title,
            description   => $description,
            playlist_url  => $playlist_url,
            thumbnail_url => $thumbnail_url,
            tags          => $tags,
            start_at      => $start_at,
            end_at        => $end_at,
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

    method title($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'title'); $title = $new };
        $title
    }

    method description($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'description'); $description = $new };
        $description
    }

    method playlist_url($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'playlist_url'); $playlist_url = $new };
        $playlist_url
    }

    method thumbnail_url($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'thumbnail_url'); $thumbnail_url = $new };
        $thumbnail_url
    }

    method start_at($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'start_at'); $start_at = $new };
        $start_at
    }

    method end_at($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'end_at'); $end_at = $new };
        $end_at
    }

    method owner($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::User'], $new, 'owner'); $owner = $new };
        $owner
    }

    method tags($new=undef) {
        if (defined $new) { assert_field(ArrayRef[InstanceOf['Isupipe::Entity::Tag']], $new, 'tags'); $tags = $new };
        $tags
    }
}
