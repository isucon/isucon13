<?php

declare(strict_types=1);

namespace IsuPipe\Stats;

use IsuPipe\AbstractHandler;
use IsuPipe\Livecomment\LivecommentModel;
use IsuPipe\Livestream\LivestreamModel;
use IsuPipe\User\{ UserModel, VerifyUserSession };
use PDO;
use PDOException;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Exception\{ HttpBadRequestException, HttpInternalServerErrorException };
use SlimSession\Helper as Session;

/**
 * @phpstan-import-type LivestreamRanking from LivestreamRankingEntry
 * @phpstan-import-type UserRanking from UserRankingEntry
 */
class Handler extends AbstractHandler
{
    use VerifyUserSession;

    public function __construct(
        private PDO $db,
        private Session $session,
    ) {
    }

    /**
     * @param array<string, string> $params
     */
    public function getUserStatisticsHandler(Request $request, Response $response, array $params): Response
    {
        $this->verifyUserSession($request, $this->session);

        $username = $params['username'] ?? '';
        // ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
        // また、現在の合計視聴者数もだす

        $this->db->beginTransaction();

        try {
            $stmt = $this->db->prepare('SELECT * FROM users WHERE name = ?');
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
            throw new HttpBadRequestException(
                request: $request,
                message: 'not found user that has the given username',
            );
        }
        $userModel = UserModel::fromRow($row);

        // ランク算出
        /** @var list<UserModel> $users */
        $users = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM users');
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $users[] = UserModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get users: ' . $e->getMessage(),
                previous: $e,
            );
        }

        /**
         * @var UserRanking $ranking
         */
        $ranking = [];
        foreach ($users as $user) {
            $query = <<<SQL
                SELECT COUNT(*) FROM users u
                INNER JOIN livestreams l ON l.user_id = u.id
                INNER JOIN reactions r ON r.livestream_id = l.id
                WHERE u.id = ?
            SQL;
            try {
                $stmt = $this->db->prepare($query);
                $stmt->bindValue(1, $user->id);
                $stmt->execute();
                $reactions = $stmt->fetchColumn();
                assert(is_int($reactions));
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to count reactions: ' . $e->getMessage(),
                    previous: $e,
                );
            }

            $query = <<<SQL
                SELECT IFNULL(SUM(l2.tip), 0) FROM users u
                INNER JOIN livestreams l ON l.user_id = u.id
                INNER JOIN livecomments l2 ON l2.livestream_id = l.id
                WHERE u.id = ?
            SQL;
            try {
                $stmt = $this->db->prepare($query);
                $stmt->bindValue(1, $user->id);
                $stmt->execute();
                $tips = (int) $stmt->fetchColumn();
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to count tips: ' . $e->getMessage(),
                    previous: $e,
                );
            }

            $score = $reactions + $tips;
            $ranking[] = new UserRankingEntry(
                username: $user->name,
                score: $score,
            );
        }
        usort($ranking, fn (UserRankingEntry $a, UserRankingEntry $b) => $b->compare($a));

        $rank = 1;
        foreach ($ranking as $entry) {
            if ($entry->username === $username) {
                break;
            }
            $rank++;
        }

        // リアクション数
        $query = <<<SQL
            SELECT COUNT(*) FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
        SQL;
        try {
            $stmt = $this->db->prepare($query);
            $stmt->bindValue(1, $username);
            $stmt->execute();
            $totalReactions = $stmt->fetchColumn();
            assert(is_int($totalReactions));
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to count total reactions: ' . $e->getMessage(),
                previous: $e,
            );
        }

        // ライブコメント数、チップ合計
        $totalLivecomments = 0;
        $totalTip = 0;
        /** @var list<LivestreamModel> $livestreams */
        $livestreams = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams WHERE user_id = ?');
            $stmt->bindValue(1, $userModel->id);
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $livestreams[] = LivestreamModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestreams: ' . $e->getMessage(),
                previous: $e,
            );
        }

        foreach ($livestreams as $livestream) {
            /** @var list<LivecommentModel> $livecomments */
            $livecomments = [];
            try {
                $stmt = $this->db->prepare('SELECT * FROM livecomments WHERE livestream_id = ?');
                $stmt->bindValue(1, $livestream->id);
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
                $totalTip += $livecomment->tip;
                $totalLivecomments++;
            }
        }

        // 合計視聴者数
        $viewersCount = 0;

        foreach ($livestreams as $livestream) {
            try {
                $stmt = $this->db->prepare('SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?');
                $stmt->bindValue(1, $livestream->id);
                $stmt->execute();
                $cnt = $stmt->fetchColumn();
                assert(is_int($cnt));
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to get livestream_view_history: ' . $e->getMessage(),
                    previous: $e,
                );
            }
            $viewersCount += $cnt;
        }

        // お気に入り絵文字
        $query = <<<SQL
            SELECT r.emoji_name
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
            GROUP BY emoji_name
            ORDER BY COUNT(*) DESC, emoji_name DESC
            LIMIT 1
        SQL;
        try {
            $stmt = $this->db->prepare($query);
            $stmt->bindValue(1, $username);
            $stmt->execute();
            $favoriteEmoji = $stmt->fetchColumn();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to find favorite emoji: ' . $e->getMessage(),
                previous: $e,
            );
        }
        if ($favoriteEmoji === false) {
            $favoriteEmoji = '';
        }
        assert(is_string($favoriteEmoji));

        $stats = new UserStatistics(
            rank: $rank,
            viewersCount: $viewersCount,
            totalReactions: $totalReactions,
            totalLivecomments: $totalLivecomments,
            totalTip: $totalTip,
            favoriteEmoji: $favoriteEmoji,
        );

        return $this->jsonResponse($response, $stats);
    }

    /**
     * @param array<string, string> $params
     */
    public function getLivestreamStatisticsHandler(Request $request, Response $response, array $params): Response
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
            throw new HttpBadRequestException(
                request: $request,
                message: 'cannot get stats of not found livestream',
            );
        }

        /** @var list<LivestreamModel> $livestreams */
        $livestreams = [];
        try {
            $stmt = $this->db->prepare('SELECT * FROM livestreams');
            $stmt->execute();
            while (($row = $stmt->fetch()) !== false) {
                $livestreams[] = LivestreamModel::fromRow($row);
            }
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to get livestreams: ' . $e->getMessage(),
                previous: $e,
            );
        }

        // ランク算出
        /** @var LivestreamRanking $ranking */
        $ranking = [];
        foreach ($livestreams as $livestream) {
            try {
                $stmt = $this->db->prepare('SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = ?');
                $stmt->bindValue(1, $livestream->id, PDO::PARAM_INT);
                $stmt->execute();
                $reactions = $stmt->fetchColumn();
                assert(is_int($reactions));
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to count reactions: ' . $e->getMessage(),
                    previous: $e,
                );
            }

            try {
                $stmt = $this->db->prepare('SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?');
                $stmt->bindValue(1, $livestream->id, PDO::PARAM_INT);
                $stmt->execute();
                $totalTips = (int) $stmt->fetchColumn();
            } catch (PDOException $e) {
                throw new HttpInternalServerErrorException(
                    request: $request,
                    message: 'failed to count tips: ' . $e->getMessage(),
                    previous: $e,
                );
            }

            $score = $reactions + $totalTips;
            $ranking[] = new LivestreamRankingEntry(
                livestreamId: $livestream->id,
                score: $score,
            );
        }
        usort($ranking, fn (LivestreamRankingEntry $a, LivestreamRankingEntry $b) => $b->compare($a));

        $rank = 1;
        foreach ($ranking as $entry) {
            if ($entry->livestreamId === $livestreamId) {
                break;
            }
            $rank++;
        }

        // 視聴者数算出
        try {
            $stmt = $this->db->prepare('SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            $viewersCount = $stmt->fetchColumn();
            assert(is_int($viewersCount));
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to count livestream viewers: ' . $e->getMessage(),
                previous: $e,
            );
        }

        // 最大チップ額
        try {
            $stmt = $this->db->prepare('SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            $maxTip = (int) $stmt->fetchColumn();
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to find maximum tip livecomment: ' . $e->getMessage(),
                previous: $e,
            );
        }

        // リアクション数
        try {
            $stmt = $this->db->prepare('SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            $totalReactions = $stmt->fetchColumn();
            assert(is_int($totalReactions));
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to count total reactions: ' . $e->getMessage(),
                previous: $e,
            );
        }

        // スパム報告数
        try {
            $stmt = $this->db->prepare('SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?');
            $stmt->bindValue(1, $livestreamId, PDO::PARAM_INT);
            $stmt->execute();
            $totalReports = $stmt->fetchColumn();
            assert(is_int($totalReports));
        } catch (PDOException $e) {
            throw new HttpInternalServerErrorException(
                request: $request,
                message: 'failed to count total spam reports: ' . $e->getMessage(),
                previous: $e,
            );
        }

        $this->db->commit();

        return $this->jsonResponse($response, new LivestreamStatistics(
            rank: $rank,
            viewersCount: $viewersCount,
            maxTip: $maxTip,
            totalReactions: $totalReactions,
            totalReports: $totalReports,
        ));
    }
}
