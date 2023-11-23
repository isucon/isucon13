use v5.38;
use experimental qw(class);

class Isupipe::Entity::User {
    field $id :param = undef;
    field $name :param = undef;
    field $display_name :param = undef;
    field $description :param = undef;
    field $password :param = undef;  # $password is hashed password
    field $theme :param = undef;
    field $icon_hash :param = undef;

    use Isupipe::Assert;
    use Types::Standard -types;

    ADJUST {
        assert_field(Int, $id, 'id');
        assert_field(Str, $name, 'name');
        assert_field(Str, $display_name, 'display_name');
        assert_field(Str, $description, 'description');
        assert_field(Str, $password, 'password');
        assert_field(InstanceOf['Isupipe::Entity::Theme'], $theme, 'theme');
        assert_field(Str, $icon_hash, 'icon_hash');
    }

    method as_hashref() {
        return {
            id           => $id,
            name         => $name,
            display_name => $display_name,
            description  => $description,
            password     => $password,
        }
    }

    method TO_JSON() {
        return {
            id => $id,
            name => $name,
            defined $display_name ? (display_name => $display_name) : (),
            defined $description ? (description => $description) : (),
            defined $theme ? (theme => $theme) : (),
            defined $icon_hash ? (icon_hash => $icon_hash) : (),
        }
    }

    method id($new=undef) {
        if (defined $new) { assert_field(Int, $new, 'id'); $id = $new };
        $id
    }

    method name($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'name'); $name = $new };
        $name
    }

    method display_name($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'display_name'); $display_name = $new };
        $display_name
    }

    method description($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'description'); $description = $new };
        $description
    }

    method password($new=undef) {
        if (defined $new) { assert_field(Str, $new, 'password'); $password = $new };
        $password
    }

    method theme($new=undef) {
        if (defined $new) { assert_field(InstanceOf['Isupipe::Entity::Theme'], $new, 'theme'); $theme = $new };
        $theme
    }
}

