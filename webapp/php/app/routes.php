<?php

declare(strict_types=1);

use IsuPipe\InitializeHandler;
use IsuPipe\Livecomment\Handler as LivecommentHandler;
use IsuPipe\Livestream\Handler as LivestreamHandler;
use IsuPipe\Payment\Handler as PaymentHandler;
use IsuPipe\Reaction\Handler as ReactionHandler;
use IsuPipe\Stats\Handler as StatsHandler;
use IsuPipe\Top\Handler as TopHandler;
use IsuPipe\User\Handler as UserHandler;
use Slim\App;

return function (App $app) {
    // 初期化
    $app->post('/api/initialize', InitializeHandler::class);

    // top
    $app->get('/api/tag', TopHandler::class . ':getTagHandler');
    $app->get('/api/user/{username}/theme', TopHandler::class . ':getStreamerThemeHandler');

    // livestream
    // reserve livestream
    $app->post('/api/livestream/reservation', LivestreamHandler::class . ':reserveLivestreamHandler');
    // list livestream
    $app->get('/api/livestream/search', LivestreamHandler::class . ':searchLivestreamsHandler');
    $app->get('/api/livestream', LivestreamHandler::class . ':getMyLivestreamsHandler');
    $app->get('/api/user/{username}/livestream', LivestreamHandler::class . ':getUserLivestreamsHandler');
    // get livestream
    $app->get('/api/livestream/{livestream_id}', LivestreamHandler::class . ':getLivestreamHandler');
    // get polling livecomment timeline
    $app->get('/api/livestream/{livestream_id}/livecomment', LivecommentHandler::class . ':getLivecommentsHandler');
    // ライブコメント投稿
    $app->post('/api/livestream/{livestream_id}/livecomment', LivecommentHandler::class . ':postLivecommentHandler');
    $app->post('/api/livestream/{livestream_id}/reaction', ReactionHandler::class . ':postReactionHandler');
    $app->get('/api/livestream/{livestream_id}/reaction', ReactionHandler::class . ':getReactionsHandler');

    // (配信者向け)ライブコメントの報告一覧取得API
    $app->get('/api/livestream/{livestream_id}/report', LivestreamHandler::class . ':getLivecommentReportsHandler');
    $app->get('/api/livestream/{livestream_id}/ngwords', LivecommentHandler::class . ':getNgwords');
    // ライブコメント報告
    $app->post('/api/livestream/{livestream_id}/livecomment/{livecomment_id}/report', LivecommentHandler::class . ':reportLivecommentHandler');
    // 配信者によるモデレーション (NGワード登録)
    $app->post('/api/livestream/{livestream_id}/moderate', LivecommentHandler::class . ':moderateHandler');

    // livestream_viewersにINSERTするため必要
    // ユーザ視聴開始 (viewer)
    $app->post('/api/livestream/{livestream_id}/enter', LivestreamHandler::class . ':enterLivestreamHandler');
    // ユーザ視聴終了 (viewer)
    $app->delete('/api/livestream/{livestream_id}/exit', LivestreamHandler::class . ':exitLivestreamHandler');

    // user
    $app->post('/api/register', UserHandler::class . ':registerHandler');
    $app->post('/api/login', UserHandler::class . ':loginHandler');
    $app->get('/api/user/me', UserHandler::class . ':getMeHandler');
    // フロントエンドで、配信予約のコラボレーターを指定する際に必要
    $app->get('/api/user/{username}', UserHandler::class . ':getUserHandler');
    $app->get('/api/user/{username}/statistics', StatsHandler::class . ':getUserStatisticsHandler');
    $app->get('/api/user/{username}/icon', UserHandler::class . ':getIconHandler');
    $app->post('/api/icon', UserHandler::class . ':postIconHandler');

    // stats
    // ライブ配信統計情報
    $app->get('/api/livestream/{livestream_id}/statistics', StatsHandler::class . ':getLivestreamStatisticsHandler');

    // 課金情報
    $app->get('/api/payment', PaymentHandler::class . ':getPaymentResult');
};
