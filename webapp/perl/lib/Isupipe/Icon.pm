package Isupipe::Icon;
use v5.38;
use utf8;

use Exporter 'import';

our @EXPORT_OK = qw(
    read_fallback_user_icon_image
);

use constant FALLBACK_IMAGE_PATH => "../img/NoImage.jpg";

sub read_fallback_user_icon_image {
    open my $fh, '<:raw', FALLBACK_IMAGE_PATH or die "Cannot open FALLBACK_IMAGE: $!";
    my $image = do { local $/; <$fh> };
    return $image;
}
