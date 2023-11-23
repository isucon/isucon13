<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use IsuPipe\Top\{ Tag, TagModel };
use IsuPipe\User\{ FillUserResponse, UserModel };
use PDO;
use RuntimeException;

trait FillLivestreamResponse
{
    use FillUserResponse;

    /**
     * @throws RuntimeException
     */
    protected function fillLivestreamResponse(LivestreamModel $livestreamModel, PDO $db): Livestream
    {
        $stmt = $db->prepare('SELECT * FROM users WHERE id = ?');
        $stmt->bindValue(1, $livestreamModel->userId, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException('not found user that has the given id');
        }

        $ownerModel = UserModel::fromRow($row);
        $owner = $this->fillUserResponse($ownerModel, $db);

        /** @var list<LivestreamTagModel> $livestreamTagModels */
        $livestreamTagModels = [];
        $stmt = $db->prepare('SELECT * FROM livestream_tags WHERE livestream_id = ?');
        $stmt->bindValue(1, $livestreamModel->id, PDO::PARAM_INT);
        $stmt->execute();
        while (($row = $stmt->fetch()) !== false) {
            $livestreamTagModels[] = LivestreamTagModel::fromRow($row);
        }

        /** @var list<Tag> $tags */
        $tags = [];
        foreach ($livestreamTagModels as $livestreamTagModel) {
            $stmt = $db->prepare('SELECT * FROM tags WHERE id = ?');
            $stmt->bindValue(1, $livestreamTagModel->tagId, PDO::PARAM_INT);
            $stmt->execute();
            $row = $stmt->fetch();
            if ($row === false) {
                throw new RuntimeException('not found tag that has the given id');
            }

            $tagModel = TagModel::fromRow($row);
            $tags[] = new Tag(
                id: $tagModel->id,
                name: $tagModel->name,
            );
        }

        return new Livestream(
            id: $livestreamModel->id,
            owner: $owner,
            title: $livestreamModel->title,
            tags: $tags,
            description: $livestreamModel->description,
            playlistUrl: $livestreamModel->playlistUrl,
            thumbnailUrl: $livestreamModel->thumbnailUrl,
            startAt: $livestreamModel->startAt,
            endAt: $livestreamModel->endAt,
        );
    }
}
