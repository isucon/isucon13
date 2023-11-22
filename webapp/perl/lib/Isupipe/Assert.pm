package Isupipe::Assert;
use v5.38;

use Exporter 'import';

our @EXPORT = qw(
    ASSERT
    assert_field
);

use Carp qw(croak);

# 本番環境ではassertしない
# 下記のassert_field以外にも、Isupipe::Util#check_paramsでも利用
use constant ASSERT => ($ENV{PLACK_ENV}||'') ne 'production';

# 開発環境では、型チェックをしてあげる
sub assert_field($type, $value, $field_name) {
    if (ASSERT && defined $value) {
        unless ($type->check($value)) {
            croak "Invalid field `$field_name`: " . $type->get_message($value);
        }
    }
}

