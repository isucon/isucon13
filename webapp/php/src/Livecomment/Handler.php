<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\AbstractHandler;
use IsuPipe\User\VerifyUserSession;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Log\LoggerInterface as Logger;
use RuntimeException;
use Slim\Exception\{ HttpBadRequestException, HttpInternalServerErrorException };
use SlimSession\Helper as Session;
use UnexpectedValueException;

class Handler extends AbstractHandler
{
    use FillLivecommentResponse, VerifyUserSession;

    public function __construct(
        private PDO $db,
        private Session $session,
        private Logger $logger,
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

    /**
     * @param array<string, string> $params
     */
    public function postLivecommentHandler(Request $request, Response $response, array $params): Response
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

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        try {
            $req = PostLivecommentRequest::fromJson($request->getBody()->getContents());
        } catch (UnexpectedValueException $e) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'failed to decode the request body as json',
                previous: $e,
            );
        }

        $this->db->beginTransaction();

        // スパム判定
        /** @var list<NGWord> $ngWords */
        $ngWords = [];
        try {
            $stmt = $this->db->prepare('SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = ? AND livestream_id = ?');
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

        foreach ($ngWords as $ngWord) {
            try {
                $query = <<<SQL
                    SELECT COUNT(*)
                    FROM
                    (SELECT ? AS text) AS texts
                    INNER JOIN
                    (SELECT CONCAT("%", ?, "%") AS pattern) AS patterns
                    ON texts.text LIKE patterns.pattern;
                SQL;
                $stmt = $this->db->prepare($query);
                $stmt->bindValue(1, $req->comment);
                $stmt->bindValue(2, $ngWord->word);
                $stmt->execute();
                $hitSpam = $stmt->fetchColumn();
                assert(is_int($hitSpam));
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get hitspam',
                    previous: $e,
                );
            }
            $this->logger->info(sprintf('[hitSpam=%d] comment = %s', $hitSpam, $req->comment));
            if ($hitSpam >= 1) {
                throw new HttpBadRequestException(
                    request: $request,
                    message: 'このコメントがスパム判定されました',
                );
            }
        }

        $now = time();
        $livecommentModel = new LivecommentModel(
            userId: $userId,
            livestreamId: $livestreamId,
            comment: $req->comment,
            tip: $req->tip,
            createdAt: $now,
        );
        try {
            $stmt = $this->db->prepare('INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (:user_id, :livestream_id, :comment, :tip, :created_at)');
            $stmt->bindValue(':user_id', $livecommentModel->userId, PDO::PARAM_INT);
            $stmt->bindValue(':livestream_id', $livecommentModel->livestreamId, PDO::PARAM_INT);
            $stmt->bindValue(':comment', $livecommentModel->comment, PDO::PARAM_STR);
            $stmt->bindValue(':tip', $livecommentModel->tip, PDO::PARAM_INT);
            $stmt->bindValue(':created_at', $livecommentModel->createdAt, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert livecomment',
                previous: $e,
            );
        }

        $livecommentId = (int) $this->db->lastInsertId();
        $livecommentModel->id = $livecommentId;

        try {
            $livecomment = $this->fillLivecommentResponse($livecommentModel, $this->db);
        } catch (RuntimeException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to fill livecomment',
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livecomment, 201);
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
