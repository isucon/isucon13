<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\AbstractHandler;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler extends AbstractHandler
{
    public function getLivecommentsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getNgwords(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function postLivecommentHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function reportLivecommentHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * NGワードを登録
     */
    public function moderateHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
