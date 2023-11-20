#!/usr/bin/perl

use strict;
use warnings;
use 5.10.0;
use List::Util;

my $MAX_TAG_ID=103;
my $MAX_LIVESTREAM_ID=7305;

# 1,2個ずつぐらいつける
# たまに5個
# すごくたまにすごく多いやつがいる

sub randomTags {
    my $num = shift;
    my @tags = List::Util::shuffle (1..$MAX_TAG_ID);
    return splice(@tags,0,$num);
}

sub genSQL {
    my $fh = shift;
    my @tag_value = @_;
    print $fh "INSERT INTO livestream_tags (id, livestream_id, tag_id) VALUES " . join(', ', @tag_value) . ";\n";
}

srand(1565458009);
open(my $sql_fh, ">:utf8", "/tmp/initial_livestream_tags.sql") or die $!;
open(my $go_fh, ">:utf8", "/tmp/livestream_tags_pool.go") or die $!;

my %tag_streams;
for my $tag_id (1..$MAX_TAG_ID) {
    $tag_streams{$tag_id} = [];
}

my @tag_value;
my $livestream_tags_id=0;
foreach my $livestream_id (1..$MAX_LIVESTREAM_ID) {
    my @tags = randomTags(int(rand(2))+1);
    if ( int(rand(10)) ==0 ) {
        # たまに多い
        my @tags = randomTags(5);
    }
    if ( int(rand(47)) ==0 ) {
        # すごくたまに多い
        my @tags = randomTags(int(rand(10))+30);
    }
    for my $tag_id (@tags) {
        $livestream_tags_id++;
        push @tag_value, sprintf('(%d, %d, %d)',$livestream_tags_id, $livestream_id, $tag_id);
        push @{$tag_streams{$tag_id}}, $livestream_id;
    }
    if (@tag_value > 500) {
        genSQL($sql_fh, @tag_value);
        @tag_value=();
    }
}

genSQL($sql_fh, @tag_value);

print $go_fh <<EOF;
package scheduler

var streamTagsPool = map[int64][]int64{
EOF

for my $tag_id (1..$MAX_TAG_ID) {
    print $go_fh "    $tag_id: []int64{";
    print $go_fh join ", ", @{$tag_streams{$tag_id}};
    print $go_fh "},\n";
}

print $go_fh <<EOF;
}

func GetStreamIDsByTagID(id int64) []int64 {
    return streamTagsPool[id]
}

EOF

