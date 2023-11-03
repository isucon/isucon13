<?php

declare(strict_types=1);

namespace IsuPipe\Top;

use IsuPipe\AbstractHandler;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\HttpInternalServerErrorException;

class Handler extends AbstractHandler
{
    public function __construct(
        private PDO $db,
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
                message: 'failed to get tags',
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
     */
    public function getStreamerThemeHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
