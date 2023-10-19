<?php

declare(strict_types=1);

namespace IsuPipe\Top;

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler
{
    public function getTagHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * 配信者のテーマ取得API
     * GET /api/user/:username/theme
     */
    public function getStreamerThemeHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
