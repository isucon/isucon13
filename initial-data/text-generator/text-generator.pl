#!/usr/bin/perl

use strict;
use warnings;
use utf8;
use Getopt::Long qw/:config posix_default no_ignore_case gnu_compat/;

srand(1565458009);

my $dic = 'positive.txt';
my $include_ngword = 0;
my $num = 20;
GetOptions(
    "dictionary|d=s" => \$dic,
    "ngword" => \$include_ngword,
    "num=i" => \$num
);

my @KEYWORDS;
{
    open(my $fh, "<:utf8", $dic) or die $!;
    @KEYWORDS = map { chomp $_; $_ } <$fh>;
}

my @NGWORDS;
{
    open(my $fh, "<:utf8", 'ngwords.txt') or die $!;
    @NGWORDS = map { chomp $_; $_ } <$fh>;
}

sub gen_text {
    my ($length, $ngword) = @_;
    my @text;
    my $ngword_index = int(rand($length));
    for (my $i=0;$i<$length;$i++) {
        my $r = int(rand(scalar @KEYWORDS));
        my $t = $KEYWORDS[$r];
        if ($i == $ngword_index && $ngword) {
            my $r = int(rand(scalar @NGWORDS));
            $t = $NGWORDS[$r];
        }
        push @text, $t;
    }
    my $text = join "", @text;
    $text =~ s/^(\s|\n)+//gs;
    $text =~ s/^(。|！|？)+//gs;
    return $text;
}

binmode(STDOUT, ":utf8");

for (1..$num) {
    my $description = gen_text(12, $include_ngword);
    print "$description\n";
}
