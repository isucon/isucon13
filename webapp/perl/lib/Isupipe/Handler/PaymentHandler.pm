package Isupipe::Handler::PaymentHandler;
use v5.38;
use utf8;

sub get_payment_result($app, $c) {

    my $total_tip = $app->dbh->select_one(
        'SELECT IFNULL(SUM(tip), 0) FROM livecomments'
    );

    return $c->render_json({
        total_tip => $total_tip,
    });
}
