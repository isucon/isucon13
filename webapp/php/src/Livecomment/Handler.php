<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\AbstractHandler;
use IsuPipe\User\VerifyUserSession;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use RuntimeException;
use Slim\Exception\{ HttpBadRequestException, HttpInternalServerErrorException };
use SlimSession\Helper as Session;

class Handler extends AbstractHandler
{
    use FillLivecommentResponse, VerifyUserSession;

    public function __construct(
        private PDO $db,
        private Session $session,
    ) {
    }

    /**
     * @param array<string, string> $params
     */
    public function getLivecommentsHandler(Request $request, Response $response, array $params): Response
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

        $query = 'SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC';
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

        /** @var list<LivecommentModel> $livecommentModels */
        $livecommentModels = [];
        try {
            $stmt = $this->db->prepare($query);
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $livecommentModels[] = LivecommentModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livecomments',
                previous: $e,
            );
        }

        /** @var list<Livecomment> $livecomments */
        $livecomments = [];
        foreach ($livecommentModels as $livecommentModel) {
            try {
                $livecomments[] = $this->fillLivecommentResponse($livecommentModel, $this->db);
            } catch (RuntimeException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to fil livecomments',
                    previous: $e,
                );
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livecomments);
    }

    /**
     * @param array<string, string> $params
     */
    public function getNgwords(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

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

        /** @var list<NGWord> $ngWords */
        $ngWords = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM ng_words WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC');
            $stmt->bindValue(1, $userId, PDO::PARAM_INT);
            $stmt->bindValue(2, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $ngWords[] = NGWord::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get NG words',
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, $ngWords);
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
