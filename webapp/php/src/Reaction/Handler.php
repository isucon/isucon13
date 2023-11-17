<?php

declare(strict_types=1);

namespace IsuPipe\Reaction;

use IsuPipe\AbstractHandler;
use IsuPipe\Livestream\{ FillLivestreamResponse, LivestreamModel };
use IsuPipe\User\{ FillUserResponse, UserModel };
use PDO;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use RuntimeException;

class Handler extends AbstractHandler
{
    use FillLivestreamResponse, FillUserResponse;

    public function getReactionsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function postReactionHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    /**
     * @throws RuntimeException
     */
    private function fillReactionResponse(ReactionModel $reactionModel, PDO $db): Reaction
    {
        $stmt = $db->prepare('SELECT * FROM users WHERE id = ?');
        $stmt->bindValue(1, $reactionModel->userId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException();
        }
        $userModel = UserModel::fromRow($row);
        $user = $this->fillUserResponse($userModel, $db);

        $stmt = $db->prepare('SELECT * FROM livestreams WHERE id = ?');
        $stmt->bindValue(1, $reactionModel->livestreamId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException();
        }
        $livestreamModel = LivestreamModel::fromRow($row);
        $livestream = $this->fillLivestreamResponse($livestreamModel, $db);

        return new Reaction(
            id: $reactionModel->id,
            emojiName: $reactionModel->emojiName,
            user: $user,
            livestream: $livestream,
            createdAt: $reactionModel->createdAt,
        );
    }
}
