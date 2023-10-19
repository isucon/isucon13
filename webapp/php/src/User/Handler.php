<?php

declare(strict_types=1);

namespace IsuPipe\User;

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler
{
    public function getIconHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function postIconHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getMeHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * ユーザ登録API
     * POST /api/register
     */
    public function registerHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * ユーザログインAPI
     * POST /api/login
     */
    public function loginHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * ユーザ詳細API
     * GET /api/user/:username
     */
    public function getUserHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
