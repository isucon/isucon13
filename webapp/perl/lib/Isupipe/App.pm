package Isupipe::App;
use v5.38;
use utf8;

use Kossy;
use HTTP::Status qw(:constants);
use DBIx::Sunny;

$Kossy::JSON_SERIALIZER = Cpanel::JSON::XS->new()->ascii(0)->utf8->convert_blessed;

use Isupipe::App::LivecommentHandler;
use Isupipe::App::LivestreamHandler;
use Isupipe::App::PaymentHandler;
use Isupipe::App::ReactionHandler;
use Isupipe::App::StatsHandler;
use Isupipe::App::TopHandler;
use Isupipe::App::UserHandler;

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
        $c->halt(HTTP_INTERNAL_SERVER_ERROR, $e);
    }

    return $c->render_json({
        advertise_level => 10,
        language        => 'perl',
    });
}

# 初期化
post '/api/initialize', \&initialize_handler;

# top
get '/api/tag', Isupipe::App::TopHandler->can('get_tag_handler');
get '/api/user/:username/theme', Isupipe::App::TopHandler->can('get_streamer_theme_handler');

# livestream
# reserve livestream
post '/api/livestream/reservation', Isupipe::App::LivestreamHandler->can('reserve_livestream_handler');
# list livestream
get '/api/livestream/search', Isupipe::App::LivestreamHandler->can('search_livestreams_handler');
get '/api/livestream', Isupipe::App::LivestreamHandler->can('get_my_livestreams_handler');
get '/api/user/:username/livestream', Isupipe::App::LivestreamHandler->can('get_user_livestreams_handler');
# get livestream
get '/api/livestream/{livestream_id:[0-9]+}', Isupipe::App::LivestreamHandler->can('get_livestream_handler');
# get polling livecomment timeline
get '/api/livestream/{livestream_id:[0-9]+}/livecomment', Isupipe::App::LivecommentHandler->can('get_livecomments_handler');
# ライブコメント投稿
post '/api/livestream/{livestream_id:[0-9]+}/livecomment',  Isupipe::App::LivecommentHandler->can('post_livecomment_handler');
# post '/livestream/:livestream_id/reaction',  \&post_reaction_handler;
# get '/livestream/:livestream_id/reaction',  \&get_reactions_handler;
#
# # (配信者向け)ライブコメント一覧取得API
# get '/livestream/:livestream_id/report',  \&get_livecomment_reports_handler;
# # ライブコメント報告
# post '/livestream/:livestream_id/livecomment/:livecomment_id/report',  \&report_livecomment_handler;
# # 配信者によるモデレーション (NGワード登録)
# post '/livestream/:livestream_id/moderate',  \&moderate_ngword_handler;
#
# # livestream_viewersにINSERTするため必要
# # ユーザ視聴開始 (viewer)
# post '/livestream/:livestream_id/enter',  \&enter_livestream_handler;
#
# # ユーザ視聴終了 (viewer)
# router 'DELETE' => '/livestream/:livestream_id/enter',  \&leave_livestream_handler;
#
# # user
#post '/user',  \&user_register_handler;
post '/api/login', Isupipe::App::UserHandler->can('login_handler');
# get '/user',  \&get_users_handler;
#
# # FIXME: ユーザ一覧を返すAPI
# # フロントエンドで、配信予約のコラボレーターを指定する際に必要
# get '/user/:user_id',  \&user_handler;
# get '/user/:user_id/theme',  \&get_user_theme_handler;
# get '/user/:user_id/statistics',  \&get_user_statistics_handler;
#
# # stats
# # ライブコメント統計情報
# get '/livestream/:livestream_id/statistics',  \&get_livestream_statistics_handler;

