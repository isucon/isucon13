<?php

declare(strict_types=1);

namespace IsuPipe\Reaction;

use IsuPipe\AbstractHandler;
use IsuPipe\Livestream\{ FillLivestreamResponse, LivestreamModel };
use IsuPipe\User\{ FillUserResponse, UserModel, VerifyUserSession };
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use RuntimeException;
use Slim\Exception\{ HttpBadRequestException, HttpInternalServerErrorException, HttpNotFoundException };
use SlimSession\Helper as Session;

class Handler extends AbstractHandler
{
    use FillLivestreamResponse, FillUserResponse, VerifyUserSession;

    public function __construct(
        private PDO $db,
        private Session $session,
    ) {
    }

    /**
     * @param array<string, string> $params
     */
    public function getReactionsHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        $livestreamIdStr = $params['livestream_id'] ?? '';
        if ($livestreamIdStr === '') {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }
        $livestreamId = filter_var($livestreamIdStr, FILTER_VALIDATE_INT);
        if (!is_int($livestreamId)) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }

        $this->db->beginTransaction();

        $query = 'SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC';
        $limitStr = $request->getQueryParams()['limit'] ?? '';
        if ($limitStr !== '') {
            $limit = filter_var($limitStr, FILTER_VALIDATE_INT);
            if (!is_int($limit)) {
                throw new HttpBadRequestException(
                    request: $request,
                    message: 'limit query parameter must be integer',
                );
            }
            $query .= sprintf(' LIMIT %d', $limit);
        }

        /** @var list<ReactionModel> $reactionModels */
        $reactionModels = [];
        try {
            $stmt = $this->db->prepare($query);
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $reactionModels[] = ReactionModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get reactions',
                previous: $e,
            );
        }

        /** @var list<Reaction> $reactions */
        $reactions = [];
        foreach ($reactionModels as $reactionModel) {
            try {
                $reactions[] = $this->fillReactionResponse($reactionModel, $this->db);
            } catch (RuntimeException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to fill reaction',
                    previous: $e,
                );
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, $reactions);
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
