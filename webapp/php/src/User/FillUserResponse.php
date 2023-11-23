<?php

declare(strict_types=1);

namespace IsuPipe\User;

use PDO;
use RuntimeException;

trait FillUserResponse
{
    protected const FALLBACK_IMAGE = __DIR__ . '/../../../img/NoImage.jpg';

    /**
     * @throws RuntimeException
     */
    protected function fillUserResponse(UserModel $userModel, PDO $db): User
    {
        $stmt = $db->prepare('SELECT * FROM themes WHERE user_id = ?');
        $stmt->bindValue(1, $userModel->id, PDO::PARAM_INT);
        $stmt->execute();
        $row = $stmt->fetch();
        if ($row === false) {
            throw new RuntimeException('not found theme that has the given user id');
        }
        $themeModel = ThemeModel::fromRow($row);

        $stmt = $db->prepare('SELECT image FROM icons WHERE user_id = ?');
        $stmt->bindValue(1, $userModel->id, PDO::PARAM_INT);
        $stmt->execute();
        $image = $stmt->fetchColumn();
        if ($image === false) {
            $image = file_get_contents($this::FALLBACK_IMAGE) ?:
                throw new RuntimeException(
                    message: 'failed to read fallback image'
                );
        }
        $iconHash = hash('sha256', $image);

        return new User(
            id: $userModel->id,
            name: $userModel->name,
            displayName: $userModel->displayName,
            description: $userModel->description,
            theme: new Theme(
                id: $themeModel->id,
                darkMode: $themeModel->darkMode,
            ),
            iconHash: $iconHash,
        );
    }
}
