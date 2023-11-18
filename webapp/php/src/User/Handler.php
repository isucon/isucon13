<?php

declare(strict_types=1);

namespace IsuPipe\User;

use App\Application\Settings\SettingsInterface as Settings;
use IsuPipe\AbstractHandler;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\{ HttpBadRequestException, HttpInternalServerErrorException, HttpNotFoundException };
use SlimSession\Helper as Session;
use UnexpectedValueException;

class Handler extends AbstractHandler
{
    use VerifyUserSession;

    const DEFAULT_USERNAME_KEY = 'USERNAME';
    const BCRYPT_DEFAULT_COST = 4;
    const FALLBACK_IMAGE = __DIR__ . '/../../../img/NoImage.jpg';

    private readonly string $powerDNSSubdomainAddress;

    public function __construct(
        private PDO $db,
        private Session $session,
        Settings $settings,
    ) {
        $this->powerDNSSubdomainAddress = $settings->get('powerdns')['subdomainAddress'];
    }

    /**
     * @param array<string, string> $params
     */
    public function getIconHandler(Request $request, Response $response, array $params): Response
    {
        $username = $params['username'] ?? '';

        $this->verifyUserSession($request, $this->session);

        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('SELECT * FROM users WHERE name = ?');
            $stmt->bindValue(1, $username);
            $stmt->execute();
            $row = $stmt->fetch();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get user',
                previous: $e,
            );
        }
        if ($row === false) {
            throw new HttpNotFoundException(
                request: $request,
                message: 'not found user that has the given username',
            );
        }
        $user = UserModel::fromRow($row);

        try {
            $stmt = $this->db->prepare('SELECT image FROM icons WHERE user_id = ?');
            $stmt->bindValue(1, $user->id);
            $stmt->execute();
            $image = $stmt->fetchColumn();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get user icon',
                previous: $e,
            );
        }
        if ($image === false) {
            $image = file_get_contents($this::FALLBACK_IMAGE) ?:
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to read fallback image'
                );
        }

        $response->getBody()->write($image);

        return $response->withHeader('Content-Type', 'image/jpeg');
    }

    public function postIconHandler(Request $request, Response $response): Response
    {
        $this->verifyUserSession($request, $this->session);

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        try {
            $req = PostIconRequest::fromJson($request->getBody()->getContents());
        } catch (UnexpectedValueException $e) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'failed to decode the request body as json',
                previous: $e,
            );
        }

        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('DELETE FROM icons WHERE user_id = ?');
            $stmt->bindValue(1, $userId, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to delete old user icon',
                previous: $e,
            );
        }

        try {
            $stmt = $this->db->prepare('INSERT INTO icons (user_id, image) VALUES (?, ?)');
            $stmt->bindValue(1, $userId, PDO::PARAM_INT);
            $stmt->bindValue(2, $req->image);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert new user icon',
                previous: $e,
            );
        }

        $iconId = (int) $this->db->lastInsertId();

        $this->db->commit();

        return $this->jsonResponse($response, new PostIconResponse(
            id: $iconId,
        ), 201);
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
