package Isupipe::App::TopHandler;
use v5.38;
use utf8;

# FIXME prefork serverで動かすからリソースの共有されない


# webappに課金サーバを兼任させる
# とりあえずfinalcheck等を実装する上で必要なので用意
my $total = 0;
my $payments = [];

sub add_payment($reservation_id, $tip) {
    push @$payments, {
        reservation_id => $reservation_id,
        tip => $tip,
    };
    $total += $tip;
}

sub get_payment_result($app, $c) {
    return $c->render_json({
        total => $total,
        payments => $payments,
    });
}
