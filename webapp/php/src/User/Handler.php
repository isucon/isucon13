<?php

declare(strict_types=1);

namespace IsuPipe\User;

use App\Application\Settings\SettingsInterface as Settings;
use IsuPipe\AbstractHandler;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;

class Handler extends AbstractHandler
{
    const DEFAULT_SESSION_EXPIRES_KEY = 'EXPIRES';
    const DEFAULT_USER_ID_KEY = 'USERID';
    const DEFAULT_USERNAME_KEY = 'USERNAME';
    const BCRYPT_DEFAULT_COST = 4;
    const FALLBACK_IMAGE = __DIR__ . '/../../../img/NoImage.jpg';

    private readonly string $powerDNSSubdomainAddress;

    public function __construct(
        Settings $settings,
    ) {
        $this->powerDNSSubdomainAddress = $settings->get('powerdns')['subdomainAddress'];
    }

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
