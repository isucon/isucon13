<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\User\{ FillUserResponse, UserModel };
use PDO;
use RuntimeException;

trait FillLivecommentReportResponse
{
    use FillLivecommentResponse;
    use FillUserResponse;

    /**
     * @throws RuntimeException
     */
    protected function fillLivecommentReportResponse(LivecommentReportModel $reportModel, PDO $db): LivecommentReport
    {
        $stmt = $db->prepare('SELECT * FROM users WHERE id = ?');
        $stmt->bindValue(1, $reportModel->userId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException('not found user that has the given id');
        }
        $reporterModel = UserModel::fromRow($row);
        $reporter = $this->fillUserResponse($reporterModel, $db);

        $stmt = $db->prepare('SELECT * FROM livecomments WHERE id = ?');
        $stmt->bindValue(1, $reportModel->livecommentId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException('not found livestream that has the given id');
        }
        $livecommentModel = LivecommentModel::fromRow($row);
        $livecomment = $this->fillLivecommentResponse($livecommentModel, $db);

        return new LivecommentReport(
            id: $reportModel->id,
            reporter: $reporter,
            livecomment: $livecomment,
            createdAt: $reportModel->createdAt,
        );
    }
}
