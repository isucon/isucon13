<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\AbstractHandler;
use IsuPipe\Livestream\LivestreamModel;
use IsuPipe\User\VerifyUserSession;
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Log\LoggerInterface as Logger;
use RuntimeException;
use Slim\Exception\{ HttpBadRequestException, HttpInternalServerErrorException, HttpNotFoundException };
use SlimSession\Helper as Session;
use UnexpectedValueException;

class Handler extends AbstractHandler
{
    use FillLivecommentResponse;
    use FillLivecommentReportResponse;
    use VerifyUserSession;

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

        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }

        $this->db->beginTransaction();

        $query = 'SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC';
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
                message: 'failed to get livecomments: ' . $e->getMessage(),
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
                    message: 'failed to fil livecomments: ' . $e->getMessage(),
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

        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
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
                message: 'failed to get NG words: ' . $e->getMessage(),
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

        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
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
                message: 'failed to decode the request body as json: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            $row = $stmt->fetch();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestream: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if ($row === false) {
            throw new HttpNotFoundException(
                request: $request,
                message: 'livestream not found',
            );
        }
        $livestreamModel = LivestreamModel::fromRow($row);

        // スパム判定
        /** @var list<NGWord> $ngWords */
        $ngWords = [];
        try {
            $stmt = $this->db->prepare('SELECT id, user_id, livestream_id, word FROM ng_words WHERE user_id = ? AND livestream_id = ?');
            $stmt->bindValue(1, $livestreamModel->userId, PDO::PARAM_INT);
            $stmt->bindValue(2, $livestreamModel->id, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $ngWords[] = NGWord::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get NG words: ' . $e->getMessage(),
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
                    message: 'failed to get hitspam: ' . $e->getMessage(),
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
                message: 'failed to insert livecomment: ' . $e->getMessage(),
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
                message: 'failed to fill livecomment: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livecomment, 201);
    }

    /**
     * @param array<string, string> $params
     */
    public function reportLivecommentHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }

        $livecommentId = $this->getAsInt($params, 'livecomment_id');
        if ($livecommentId === false) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livecomment_id in path must be integer',
            );
        }

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            $row = $stmt->fetch();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestream: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if ($row === false) {
            throw new HttpNotFoundException(
                request: $request,
                message: 'livestream not found',
            );
        }

        try {
            $stmt = $this->db->prepare('SELECT * FROM livecomments WHERE id = ?');
            $stmt->bindValue(1, $livecommentId, PDO::PARAM_INT);
            $stmt->execute();
            $row = $stmt->fetch();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livecomment: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if ($row === false) {
            throw new HttpNotFoundException(
                request: $request,
                message: 'livecomment not found',
            );
        }

        $now = time();
        $reportModel = new LivecommentReportModel(
            userId: $userId,
            livestreamId: $livestreamId,
            livecommentId: $livecommentId,
            createdAt: $now,
        );
        try {
            $stmt = $this->db->prepare('INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (:user_id, :livestream_id, :livecomment_id, :created_at)');
            $stmt->bindValue(':user_id', $reportModel->userId, PDO::PARAM_INT);
            $stmt->bindValue(':livestream_id', $reportModel->livestreamId, PDO::PARAM_INT);
            $stmt->bindValue(':livecomment_id', $reportModel->livecommentId, PDO::PARAM_INT);
            $stmt->bindValue(':created_at', $reportModel->createdAt, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert livecomment report: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $reportId = (int) $this->db->lastInsertId();
        $reportModel->id = $reportId;

        try {
            $report = $this->fillLivecommentReportResponse($reportModel, $this->db);
        } catch (RuntimeException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to fill livecomment report: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, $report, 201);
    }

    /**
     * NGワードを登録
     *
     * @param array<string, string> $params
     */
    public function moderateHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        $livestreamId = $this->getAsInt($params, 'livestream_id');
        if ($livestreamId === false) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id in path must be integer',
            );
        }

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        try {
            $req = ModerateRequest::fromJson($request->getBody()->getContents());
        } catch (UnexpectedValueException $e) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'failed to decode the request body as json: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->beginTransaction();

        // 配信者自身の配信に対するmoderateなのかを検証
        /** @var list<LivestreamModel> $ownedLivestreams */
        $ownedLivestreams = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE id = ? AND user_id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->bindValue(2, $userId, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $ownedLivestreams[] = LivestreamModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestreams: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if (count($ownedLivestreams) === 0) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'A streamer can\'t moderate livestreams that other streamers own',
            );
        }

        try {
            $stmt = $this->db->prepare('INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (:user_id, :livestream_id, :word, :created_at)');
            $stmt->bindValue(':user_id', $userId, PDO::PARAM_INT);
            $stmt->bindValue(':livestream_id', $livestreamId, PDO::PARAM_INT);
            $stmt->bindValue(':word', $req->ngWord);
            $stmt->bindValue(':created_at', time(), PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert new NG word: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $wordId = (int) $this->db->lastInsertId();

        /** @var list<NGWord> $ngwords */
        $ngwords = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM ng_words WHERE livestream_id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $ngwords[] = NGWord::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get NG words: ' . $e->getMessage(),
                previous: $e,
            );
        }

        // NGワードにヒットする過去の投稿も全削除する
        foreach ($ngwords as $ngword) {
            // ライブコメント一覧取得
            /** @var list<LivecommentModel> $livecomments */
            $livecomments = [];
            try {
                $stmt = $this->db->query('SELECT * FROM livecomments');
                $stmt->execute();
                while (($row = $stmt->fetch()) !== false) {
                    $livecomments[] = LivecommentModel::fromRow($row);
                }
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get livecomments: ' . $e->getMessage(),
                    previous: $e,
                );
            }

            foreach ($livecomments as $livecomment) {
                $query = <<<SQL
                    DELETE FROM livecomments
                    WHERE
                    id = ? AND
                    livestream_id = ? AND
                    (SELECT COUNT(*)
                    FROM
                    (SELECT ? AS text) AS texts
                    INNER JOIN
                    (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
                    ON texts.text LIKE patterns.pattern) >= 1;
                SQL;
                try {
                    $stmt = $this->db->prepare($query);
                    $stmt->bindValue(1, $livecomment->id, PDO::PARAM_INT);
                    $stmt->bindValue(2, $livecomment->livestreamId, PDO::PARAM_INT);
                    $stmt->bindValue(3, $livecomment->comment);
                    $stmt->bindValue(4, $ngword->word);
                    $stmt->execute();
                } catch (PDOException $e) {
                    throw new HttpInternalServerErrorException(
                        request: $request,
                        message: 'failed to delete old livecomments that hit spams: ' . $e->getMessage(),
                        previous: $e,
                    );
                }
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, ['word_id' => $wordId], 201);
    }
}
