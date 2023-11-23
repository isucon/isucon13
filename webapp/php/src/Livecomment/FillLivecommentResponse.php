<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\Livestream\{ FillLivestreamResponse, LivestreamModel };
use IsuPipe\User\{ FillUserResponse, UserModel };
use PDO;
use RuntimeException;

trait FillLivecommentResponse
{
    use FillLivestreamResponse;
    use FillUserResponse;

    /**
     * @throws RuntimeException
     */
    protected function fillLivecommentResponse(LivecommentModel $livecommentModel, PDO $db): Livecomment
    {
        $stmt = $db->prepare('SELECT * FROM users WHERE id = ?');
        $stmt->bindValue(1, $livecommentModel->userId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException('not found user that has the given id');
        }
        $commentOwnerModel = UserModel::fromRow($row);
        $commentOwner = $this->fillUserResponse($commentOwnerModel, $db);

        $stmt = $db->prepare('SELECT * FROM livestreams WHERE id = ?');
        $stmt->bindValue(1, $livecommentModel->livestreamId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException('not found livestream that has the given id');
        }
        $livestreamModel = LivestreamModel::fromRow($row);
        $livestream = $this->fillLivestreamResponse($livestreamModel, $db);

        return new Livecomment(
            id: $livecommentModel->id,
            user: $commentOwner,
            livestream: $livestream,
            comment: $livecommentModel->comment,
            tip: $livecommentModel->tip,
            createdAt: $livecommentModel->createdAt,
        );
    }
}
