package Isupipe::Assert;
use v5.38;

use Exporter 'import';

our @EXPORT = qw(
    ASSERT
    assert_field
);

use Carp qw(croak);

use constant ASSERT => ($ENV{PLACK_ENV}||'') ne 'production';

sub assert_field($type, $value, $field_name) {
    if (ASSERT && defined $value) {
        unless ($type->check($value)) {
            croak "Invalid field `$field_name`: " . $type->get_message($value);
        }
    }
}

