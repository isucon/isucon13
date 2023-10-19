<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler
{
    public function reserveLivestreamHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function searchLivestreamsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getMyLivestreamsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getUserLivestreamsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * viewerテーブルの廃止
     */
    public function enterLivestreamHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function exitLivestreamHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getLivestreamHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getLivecommentReportsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
