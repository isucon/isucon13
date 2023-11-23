package Isupipe::App;
use v5.38;
use utf8;
use experimental qw(try);

use Kossy;
use HTTP::Status qw(:constants);
use DBIx::Sunny;

$Kossy::JSON_SERIALIZER = Cpanel::JSON::XS->new()->ascii(0)->utf8->convert_blessed;

use Isupipe::Log;
use Isupipe::Handler::LivecommentHandler;
use Isupipe::Handler::LivestreamHandler;
use Isupipe::Handler::PaymentHandler;
use Isupipe::Handler::ReactionHandler;
use Isupipe::Handler::StatsHandler;
use Isupipe::Handler::TopHandler;
use Isupipe::Handler::UserHandler;

sub connect_db() {
    my $host     = $ENV{ISUCON13_MYSQL_DIALCONFIG_ADDRESS}    || '127.0.0.1';
    my $port     = $ENV{ISUCON13_MYSQL_DIALCONFIG_PORT}       || '3306';
    my $user     = $ENV{ISUCON13_MYSQL_DIALCONFIG_USER}       || 'isucon';
    my $password = $ENV{ISUCON13_MYSQL_DIALCONFIG_PASSWORD}   || 'isucon';
    my $dbname   = $ENV{ISUCON13_MYSQL_DIALCONFIG_DATABASE}   || 'isupipe';

    my $dsn = "dbi:mysql:database=$dbname;host=$host;port=$port";
    my $dbh = DBIx::Sunny->connect($dsn, $user, $password, {
        mysql_enable_utf8mb4 => 1,
        mysql_auto_reconnect => 1,
    });
    return $dbh;
}

sub dbh($self) {
    $self->{_dbh} //= connect_db();
}

sub initialize_handler($self, $c) {
    my $e = system($self->root_dir . "/../sql/init.sh");
    if ($e) {
        warnf("init.sh failed with err=%s", $e);
        $c->halt(HTTP_INTERNAL_SERVER_ERROR, "faild to initialize: $e");
    }

    return $c->render_json({
        language        => 'perl',
    });
}

sub h($klass, $name) {
    my $handler_class = "Isupipe::Handler::$klass";
    my $handler = $handler_class->can($name);
    unless ($handler) {
        local $Log::Minimal::TRACE_LEVEL = $Log::Minimal::TRACE_LEVEL + 1;
        croakf("handler `%s` not found in %s", $name, $handler_class);
    }
    return sub ($app, $c) {
        try {
            my $response = $handler->($app, $c);
            return $response;
        } catch ($error) {
            error_response_handler($error, $app, $c);
        }
    }
}

# 初期化
post '/api/initialize', \&initialize_handler;

# top
get '/api/tag', h(TopHandler => 'get_tag_handler');

# livestream
# reserve livestream
post '/api/livestream/reservation', h(LivestreamHandler => 'reserve_livestream_handler');
# list livestream
get '/api/livestream/search', h(LivestreamHandler => 'search_livestreams_handler');
get '/api/livestream', h(LivestreamHandler => 'get_my_livestreams_handler');

# get livestream
get '/api/livestream/{livestream_id:[0-9]+}', h(LivestreamHandler => 'get_livestream_handler');

# get polling livecomment timeline
get '/api/livestream/{livestream_id:[0-9]+}/livecomment', h(LivecommentHandler => 'get_livecomments_handler');
# ライブコメント投稿
post '/api/livestream/{livestream_id:[0-9]+}/livecomment', h(LivecommentHandler => 'post_livecomment_handler');
post '/api/livestream/{livestream_id:[0-9]+}/reaction', h(ReactionHandler => 'post_reaction_handler');
get  '/api/livestream/{livestream_id:[0-9]+}/reaction', h(ReactionHandler => 'get_reactions_handler');

# (配信者向け)ライブコメント一覧取得API
get '/api/livestream/{livestream_id:[0-9]+}/report', h(LivestreamHandler => 'get_livecomment_reports_handler');
get '/api/livestream/{livestream_id:[0-9]+}/ngwords', h(LivecommentHandler => 'get_ngwords_handler');

# ライブコメント報告
post '/api/livestream/{livestream_id:[0-9]+}/livecomment/:livecomment_id/report',  h(LivecommentHandler => 'report_livecomment_handler');
# 配信者によるモデレーション (NGワード登録)
post '/api/livestream/{livestream_id:[0-9]+}/moderate', h(LivecommentHandler => 'moderate_handler');

# livestream_viewersにINSERTするため必要
# ユーザ視聴開始 (viewer)
post '/api/livestream/{livestream_id:[0-9]+}/enter', h(LivestreamHandler => 'enter_livestream_handler');
# ユーザ視聴終了 (viewer)
router 'DELETE' => '/api/livestream/{livestream_id:[0-9]+}/exit',  h(LivestreamHandler => 'exit_livestream_handler');

# user
post '/api/register', h(UserHandler => 'register_handler');
post '/api/login', h(UserHandler => 'login_handler');
get '/api/user/me',  h(UserHandler => 'get_me_handler');

# フロントエンドで、配信予約のコラボレーターを指定する際に必要
get '/api/user/:username',  h(UserHandler => 'get_user_handler');
get '/api/user/:username/livestream', h(LivestreamHandler => 'get_user_livestreams_handler');
get '/api/user/:username/theme', h(TopHandler => 'get_streamer_theme_handler');
get '/api/user/:username/statistics',  h(StatsHandler => 'get_user_statistics_handler');
get '/api/user/:username/icon',  h(UserHandler => 'get_icon_handler');
post '/api/icon',  h(UserHandler => 'post_icon_handler');

# stats
# ライブ配信統計情報
get '/api/livestream/{livestream_id:[0-9]+}/statistics', h(StatsHandler => 'get_livestream_statistics_handler');

# 課金情報
get '/api/payment', h(PaymentHandler => 'get_payment_result');


sub error_response_handler($error, $app, $c) {
    if ($error isa Kossy::Exception) {
        if ($error->{response}) {
            die $error; # rethrow
        }

        # JSON にして投げ直し
        debugf("(Kossy::Exception) %s %s%s : %s %s", $c->req->method, $c->env->{HTTP_HOST}, $c->req->path, $error->{code}, $error->{message});

        my $res = $c->render_json({
            error => $error->{message},
        });
        $res->status($error->{code});
        return $res;
    }

    warnf("error at %s: %s", $c->req->path, $error);

    my $res = $c->render_json({
        error => Isupipe::Log::ddf($error),
    });
    $res->status(HTTP_INTERNAL_SERVER_ERROR);
    return $res;
}
