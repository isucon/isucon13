<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use App\Application\Settings\SettingsInterface as Settings;
use DateTimeImmutable;
use DateTimeZone;
use ErrorException;
use IsuPipe\AbstractHandler;
use IsuPipe\User\{ UserModel, VerifyUserSession };
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
    use FillLivestreamResponse, VerifyUserSession;

    private readonly int $numReservationSlot;

    /**
     * @throws ErrorException
     */
    public function __construct(
        private PDO $db,
        private Session $session,
        private Logger $logger,
        Settings $settings,
    ) {
        $this->numReservationSlot = $settings->get('numReservationSlot');
    }

    public function reserveLivestreamHandler(Request $request, Response $response): Response
    {
        $this->verifyUserSession($request, $this->session);

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        try {
            $req = ReserveLivestreamRequest::fromJson($request->getBody()->getContents());
        } catch (UnexpectedValueException $e) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'failed to decode the request body as json',
                previous: $e,
            );
        }

        $this->db->beginTransaction();

        // 2024/04/01からの１年間の期間内であるかチェック
        $termStartAt = DateTimeImmutable::createFromFormat('Y-m-d H:i:s', '2024-04-01 00:00:00', new DateTimeZone('UTC'));
        $termEndAt = DateTimeImmutable::createFromFormat('Y-m-d H:i:s', '2025-04-01 00:00:00', new DateTimeZone('UTC'));
        $reserveStartAt = DateTimeImmutable::createFromFormat('U', (string) $req->startAt, new DateTimeZone('UTC'));
        $reserveEndAt = DateTimeImmutable::createFromFormat('U', (string) $req->endAt, new DateTimeZone('UTC'));
        if ($reserveStartAt >= $termEndAt || $reserveEndAt <= $termStartAt) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'bad reservation time range',
            );
        }

        // 予約枠をみて、予約が可能か調べる
        /** @var list<ReservationSlotModel> $slots */
        $slots = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE');
            $stmt->bindValue(1, $req->startAt, PDO::PARAM_INT);
            $stmt->bindValue(2, $req->endAt, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $slots[] = ReservationSlotModel::fromRow($row);
            }
        } catch (PDOException $e) {
            $this->logger->warning('予約枠一覧取得でエラー発生: ' . $e->getMessage());
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get reservation_slots',
                previous: $e,
            );
        }
        foreach ($slots as $slot) {
            try {
                $stmt = $this->db->prepare('SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?');
                $stmt->bindValue(1, $slot->startAt, PDO::PARAM_INT);
                $stmt->bindValue(2, $slot->endAt, PDO::PARAM_INT);
                $stmt->execute();
                $count = $stmt->fetchColumn();
                assert(is_int($count));
                $this->logger->info(sprintf('%d ~ %d予約枠の残数 = %d', $slot->startAt, $slot->endAt, $slot->slot));
                if ($count < 1) {
                    throw new HttpBadRequestException(
                        request: $request,
                        message: sprintf('予約区間 %d ~ %dが予約できません', $req->startAt, $req->endAt),
                    );
                }
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get reservation_slots',
                    previous: $e,
                );
            }
        }

        $livestreamModel = new LivestreamModel(
            userId: $userId,
            title: $req->title,
            description: $req->description,
            playlistUrl: $req->playlistUrl,
            thumbnailUrl: $req->thumbnailUrl,
            startAt: $req->startAt,
            endAt: $req->endAt,
        );

        try {
            $stmt = $this->db->prepare('UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?');
            $stmt->bindValue(1, $req->startAt, PDO::PARAM_INT);
            $stmt->bindValue(2, $req->endAt, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to update reservation_slots',
                previous: $e,
            );
        }

        try {
            $stmt = $this->db->prepare('INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at)');
            $stmt->bindValue(':user_id', $livestreamModel->userId, PDO::PARAM_INT);
            $stmt->bindValue(':title', $livestreamModel->title);
            $stmt->bindValue(':description', $livestreamModel->description);
            $stmt->bindValue(':playlist_url', $livestreamModel->playlistUrl);
            $stmt->bindValue(':thumbnail_url', $livestreamModel->thumbnailUrl);
            $stmt->bindValue(':start_at', $livestreamModel->startAt, PDO::PARAM_INT);
            $stmt->bindValue(':end_at', $livestreamModel->endAt, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert livestream',
                previous: $e,
            );
        }

        $livestreamId = (int) $this->db->lastInsertId();
        $livestreamModel->id = $livestreamId;

        // タグ追加
        foreach ($req->tags as $tagId) {
            try {
                $stmt = $this->db->prepare('INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (:livestream_id, :tag_id)');
                $stmt->bindValue(':livestream_id', $livestreamId, PDO::PARAM_INT);
                $stmt->bindValue(':tag_id', $tagId, PDO::PARAM_INT);
                $stmt->execute();
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to insert livestream tag',
                    previous: $e,
                );
            }
        }

        try {
            $livestream = $this->fillLivestreamResponse($livestreamModel, $this->db);
        } catch (RuntimeException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to fill livestream',
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livestream, 201);
    }

    public function searchLivestreamsHandler(Request $request, Response $response): Response
    {
        $keyTagName = $request->getQueryParams()['tag'] ?? '';

        $this->db->beginTransaction();

        /** @var list<LivestreamModel> $livestreamModels */
        $livestreamModels = [];
        if ($keyTagName !== '') {
            // タグによる取得
            try {
                $stmt = $this->db->prepare('SELECT id FROM tags WHERE name = ?');
                $stmt->bindValue(1, $keyTagName);
                $stmt->execute();
                /** @var list<int> $tagIdList */
                $tagIdList = $stmt->fetchAll(PDO::FETCH_COLUMN);
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get tags',
                    previous: $e,
                );
            }

            /** @var list<LivestreamTagModel> $keyTaggedLivestreams */
            $keyTaggedLivestreams = [];
            try {
                $sql = 'SELECT * FROM livestream_tags WHERE tag_id IN (';
                $sql .= implode(',', array_fill(0, count($tagIdList), '?'));
                $sql .= ')';
                $stmt = $this->db->prepare($sql);
                foreach ($tagIdList as $i => $tagId) {
                    $stmt->bindValue($i + 1, $tagId, PDO::PARAM_INT);
                }
                $stmt->execute();
                while (($row = $stmt->fetch()) !== false) {
                    $keyTaggedLivestreams[] = LivestreamTagModel::fromRow($row);
                }
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get keyTaggedLivestreams',
                    previous: $e,
                );
            }

            foreach ($keyTaggedLivestreams as $keyTaggedLivestream) {
                try {
                    $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE id = ?');
                    $stmt->bindValue(1, $keyTaggedLivestream->livestreamId, PDO::PARAM_INT);
                    $stmt->execute();
                    $row = $stmt->fetch();
                } catch (PDOException $e) {
                    throw new HttpInternalServerErrorException(
                        request: $request,
                        message: 'failed to get livestreams',
                        previous: $e,
                    );
                }
                if ($row === false) {
                    throw new HttpInternalServerErrorException(
                        request: $request,
                        message: 'failed to get livestreams',
                    );
                }
                $livestreamModels[] = LivestreamModel::fromRow($row);
            }
        } else {
            // 検索条件なし
            $query = 'SELECT * FROM livestreams';
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
            try {
                $stmt = $this->db->query($query);
                while (($row = $stmt->fetch()) !== false) {
                    $livestreamModels[] = LivestreamModel::fromRow($row);
                }
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get livestreams',
                    previous: $e,
                );
            }
        }

        /** @var list<Livestream> $livestreams */
        $livestreams = [];
        foreach ($livestreamModels as $livestreamModel) {
            try {
                $livestreams[] = $this->fillLivestreamResponse($livestreamModel, $this->db);
            } catch (RuntimeException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to fill livestream',
                    previous: $e,
                );
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livestreams);
    }

    public function getMyLivestreamsHandler(Request $request, Response $response): Response
    {
        $this->verifyUserSession($request, $this->session);

        $this->db->beginTransaction();

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        /** @var list<LivestreamModel> $livestreamModels */
        $livestreamModels = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE user_id = ?');
            $stmt->bindValue(1, $userId, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $livestreamModels[] = LivestreamModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestreams',
                previous: $e,
            );
        }

        /** @var list<Livestream> $livestreams */
        $livestreams = [];
        foreach ($livestreamModels as $livestreamModel) {
            try {
                $livestreams[] = $this->fillLivestreamResponse($livestreamModel, $this->db);
            } catch (RuntimeException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to fill livestream',
                    previous: $e,
                );
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livestreams);
    }

    /**
     * @param array<string, string> $params
     */
    public function getUserLivestreamsHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        $username = $params['username'] ?? '';

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
                message: 'failed to get user',
            );
        }
        $user = UserModel::fromRow($row);

        /** @var list<LivestreamModel> $livestreamModels */
        $livestreamModels = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE user_id = ?');
            $stmt->bindValue(1, $user->id, PDO::PARAM_INT);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $livestreamModels[] = LivestreamModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestreams',
                previous: $e,
            );
        }

        /** @var list<Livestream> $livestreams */
        $livestreams = [];
        foreach ($livestreamModels as $livestreamModel) {
            try {
                $livestreams[] = $this->fillLivestreamResponse($livestreamModel, $this->db);
            } catch (RuntimeException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to fill livestream',
                    previous: $e,
                );
            }
        }

        $this->db->commit();

        return $this->jsonResponse($response, $livestreams);
    }

    /**
     * viewerテーブルの廃止
     *
     * @param array<string, string> $params
     */
    public function enterLivestreamHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        // existence already checked
        $userId = $this->session->get($this::DEFAULT_USER_ID_KEY);

        $livestreamIdStr = $params['livestream_id'] ?? '';
        if ($livestreamIdStr === '') {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id must be integer',
            );
        }
        $livestreamId = filter_var($livestreamIdStr, FILTER_VALIDATE_INT);
        if (!is_int($livestreamId)) {
            throw new HttpBadRequestException(
                request: $request,
                message: 'livestream_id must be integer',
            );
        }

        $this->db->beginTransaction();

        $viewer = new LivestreamViewerModel(
            userId: $userId,
            livestreamId: $livestreamId,
            createdAt: time(),
        );

        try {
            $stmt = $this->db->prepare('INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(:user_id, :livestream_id, :created_at)');
            $stmt->bindValue(':user_id', $viewer->userId, PDO::PARAM_INT);
            $stmt->bindValue(':livestream_id', $viewer->livestreamId, PDO::PARAM_INT);
            $stmt->bindValue(':created_at', $viewer->createdAt, PDO::PARAM_INT);
            $stmt->execute();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to insert livestream_view_history',
                previous: $e,
            );
        }

        $this->db->commit();

        return $response;
    }

    public function exitLivestreamHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getLivestreamHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }

    public function getLivecommentReportsHandler(Request $request, Response $response): Response
    {
        // TODO: 実装
        return $response;
    }
}
