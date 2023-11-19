<?php

declare(strict_types=1);

namespace IsuPipe\Top;

use IsuPipe\AbstractHandler;
use IsuPipe\User\{ Theme, ThemeModel, UserModel, VerifyUserSession };
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\{ HttpInternalServerErrorException, HttpNotFoundException };
use SlimSession\Helper as Session;

class Handler extends AbstractHandler
{
    use VerifyUserSession;

    public function __construct(
        private PDO $db,
        private Session $session,
    ) {
    }

    public function getTagHandler(Request $request, Response $response): Response
    {
        $this->db->beginTransaction();

        /** @var list<TagModel> $tagModels */
        $tagModels = [];
        try {
            $stmt = $this->db->query('SELECT * FROM tags');
            while (($row = $stmt->fetch()) !== false) {
                $tagModels[] = TagModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get tags: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->commit();

        /** @var list<Tag> $tags */
        $tags = [];
        foreach ($tagModels as $tagModel) {
            $tags[] = new Tag(
                id: $tagModel->id,
                name: $tagModel->name,
            );
        }

        return $this->jsonResponse($response, new TagsResponse(
            tags: $tags
        ));
    }

    /**
     * 配信者のテーマ取得API
     * GET /api/user/:username/theme
     *
     * @param array<string, string> $params
     */
    public function getStreamerThemeHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        $username = $params['username'] ?? '';

        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('SELECT id FROM users WHERE name = ?');
            $stmt->bindValue(1, $username);
            $stmt->execute();
            $row = $stmt->fetch();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get user: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if ($row === false) {
            throw new HttpNotFoundException(
                request: $request,
                message: 'not found user that has the given username',
            );
        }
        $userModel = UserModel::fromRow($row);

        try {
            $stmt = $this->db->prepare('SELECT * FROM themes WHERE user_id = ?');
            $stmt->bindValue(1, $userModel->id, PDO::PARAM_INT);
            $stmt->execute();
            $row = $stmt->fetch();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get user theme: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if ($row === false) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get user theme',
            );
        }
        $themeModel = ThemeModel::fromRow($row);

        $this->db->commit();

        $theme = new Theme(
            id: $themeModel->id,
            darkMode: $themeModel->darkMode,
        );

        return $this->jsonResponse($response, $theme);
    }
}
