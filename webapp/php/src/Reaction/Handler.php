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
use UnexpectedValueException;

class Handler extends AbstractHandler
{
    use FillLivestreamResponse;
    use FillUserResponse;
    use VerifyUserSession;

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

        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }

        $this->db->beginTransaction();

        $query = 'SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC';
        if (isset($request->getQueryParams()['limit'])) {
            $limit = $this->getAsInt($request->getQueryParams(), 'limit');
            if ($limit === false) {
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
                message: 'failed to get reactions: ' . $e->getMessage(),
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
                    message: 'failed to fill reaction: ' . $e->getMessage(),
                    previous: $e,
                );
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, $reactions);
    }

    /**
     * @param array<string, string> $params
     */
    public function postReactionHandler(Request $request, Response $response, array $params): Response
    {
        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }

        $this->verifyUserSession($request, $this->session);

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        try {
            $req = PostReactionRequest::fromJson($request->getBody()->getContents());
        } catch (UnexpectedValueException $e) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'failed to decode the request body as json: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->beginTransaction();

        $reactionModel = new ReactionModel(
            userId: $userId,
            livestreamId: $livestreamId,
            emojiName: $req->emojiName,
            createdAt: time(),
        );

        try {
            $stmt = $this->db->prepare('INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (:user_id, :livestream_id, :emoji_name, :created_at)');
            $stmt->bindValue(':user_id', $reactionModel->userId, PDO::PARAM_INT);
            $stmt->bindValue(':livestream_id', $reactionModel->livestreamId, PDO::PARAM_INT);
            $stmt->bindValue(':emoji_name', $reactionModel->emojiName);
            $stmt->bindValue(':created_at', $reactionModel->createdAt, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert reaction: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $reactionId = (int) $this->db->lastInsertId();
        $reactionModel->id = $reactionId;

        try {
            $reaction = $this->fillReactionResponse($reactionModel, $this->db);
        } catch (RuntimeException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to fill reaction: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, $reaction, 201);
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
