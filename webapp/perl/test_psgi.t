use v5.38;
use utf8;
use lib 'lib';

# あとで消す

use Test2::V0;
use Plack::Test;
use HTTP::Request::Common;
use HTTP::Status qw(:constants);
use HTTP::Cookies;
use Cpanel::JSON::XS ();
use Cpanel::JSON::XS::Type;
use Encode ();
use Plack::Util ();

my $app = Plack::Util::load_psgi('./app.psgi');

sub decode_json {
    state $json = Cpanel::JSON::XS->new()->utf8;
    return $json->decode(@_);
}

sub with_json_request($req, $data) {
    state $json = Cpanel::JSON::XS->new->ascii(0)->utf8;
    my $encocded_json = $json->encode($data);

    $req->header('Content-Type' => 'application/json; charset=utf-8');
    $req->header('Content-Length' => length $encocded_json);
    $req->content($encocded_json);

    return $req;
}

sub login_default($cb, $req) {
    my $login_req = POST "/api/login";
    with_json_request($login_req, {
        name     => 'test001',
        password => 'test',
    });
    my $login_res = $cb->($login_req);
    if ($login_res->code != HTTP_OK) {
        die 'Failed to login: ' . Encode::encode_utf8($login_res->content);
    }

    my $cookie = $login_res->header('Set-Cookie');

    $req->header('Cookie' => $cookie);
    return $req;
}


subtest 'POST /api/initialize' => sub {
    test_psgi $app, sub ($cb){
        my $res = $cb->(POST "/api/initialize");
        is decode_json($res->content), {
            advertise_level => 10,
            language        => 'perl',
        };
    };
};

subtest 'GET /api/tag' => sub {
    test_psgi $app, sub ($cb){
        my $res = $cb->(GET "/api/tag");
        is $res->code, HTTP_OK;

        my $tags = decode_json($res->content, my $decode_type);
        is $decode_type, {
            tags => array {
                all_items {
                    id   => JSON_TYPE_INT,
                    name => JSON_TYPE_STRING,
                };
                etc;
            }
        };
    };
};

subtest 'GET /api/user/:username/theme' => sub {
    test_psgi $app, sub ($cb) {
        my $req = GET "/api/user/test001/theme";
        login_default($cb, $req);

        my $res = $cb->($req);
        is $res->code, HTTP_OK;

        is decode_json($res->content), {
            id        => 1,
            dark_mode => 0,
        };
    };
};

subtest 'POST /api/livestream/reservation' => sub {
    test_psgi $app, sub ($cb) {
        my $req = POST "/api/livestream/reservation";
        login_default($cb, $req);

        with_json_request($req, {
            tags => [43], # DIY
            title => '月曜大工',
            description => 'キーボードをつくります',
            collaborators => [],
            start_at    => 1714521600, # 2024/05/01 UTC
            end_at      => 1717200000, # 2024/06/01 UTC
        });

        my $res = $cb->($req);
        is ($res->code, HTTP_CREATED) or diag $res->content;

        is decode_json($res->content), hash {
            field id => 2;
            etc;
        };
    };
};

subtest 'GET /api/livestream/search' => sub {

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/livestream/search?tag=DIY";

        my $res = $cb->($req);
        is($res->code, HTTP_OK) or do { diag $req->as_string, $res->content };

        is decode_json($res->content), array {
            all_items hash {
                field title => D;
                etc;
            };
            etc;
        };
    };

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/livestream/search";

        my $res = $cb->($req);
        is($res->code, HTTP_OK) or do { diag $req->as_string, $res->content };

        is decode_json($res->content), array {
            all_items hash {
                field title => D;
                etc;
            };
            etc;
        };
    };

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/livestream/search?limit=1";

        my $res = $cb->($req);
        is($res->code, HTTP_OK) or do { diag $req->as_string, $res->content };

        is decode_json($res->content), array {
            all_items hash {
                field title => D;
                etc;
            };
            etc;
        };
    };
};

subtest 'GET /api/livestream' => sub {

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/livestream";
        login_default($cb, $req);

        my $res = $cb->($req);
        is $res->code, HTTP_OK;

        is decode_json($res->content), array {
            all_items hash {
                field title => D;
                etc;
            };
            etc;
        };
    };
};

subtest 'GET /api/user/:username/livestream' => sub {

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/user/test001/livestream";
        login_default($cb, $req);

        my $res = $cb->($req);
        is $res->code, HTTP_OK;

        is decode_json($res->content), array {
            all_items hash {
                field title => D;
                etc;
            };
            etc;
        };
    };
};

subtest 'GET /api/livestream/:livestream_id' => sub {

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/livestream/2";
        login_default($cb, $req);

        my $res = $cb->($req);
        is $res->code, HTTP_OK;

        is decode_json($res->content), hash {
            field title => '月曜大工';
            etc;
        };
    };
};

subtest 'GET /api/livestream/:livestream_id/livecomment' => sub {

    test_psgi $app, sub ($cb) {
        my $req = GET "/api/livestream/1/livecomment?limit=5";
        login_default($cb, $req);

        my $res = $cb->($req);
        is ($res->code, HTTP_OK) or diag $res->content;

        is decode_json($res->content), array {
            item 0 => hash {
                field id => 1;
                etc;
            };
            etc;
        };
    };
};

subtest 'POST /api/login' => sub {
    test_psgi $app, sub ($cb) {

        my $req = POST "/api/login";
        with_json_request($req, {
            name     => 'test001',
            password => 'test',
        });

        my $res = $cb->($req);
        is $res->code, HTTP_OK;
        is $res->content, '';
    };
};

subtest 'POST /api/livestream/:livestream_id/livecomment' => sub {

    test_psgi $app, sub ($cb) {
        my $req = POST "/api/livestream/1/livecomment";
        login_default($cb, $req);

        with_json_request($req, {
            comment => '応援しています!!',
            tip => 999,
        });

        my $res = $cb->($req);
        is ($res->code, HTTP_CREATED) or diag $res->content;

        is decode_json($res->content), hash {
            field tip => 999;
            field comment => '応援しています!!';
            field user => hash {
                field name => 'test001';
                etc;
            };
            etc;
        };
    };
};

done_testing;
