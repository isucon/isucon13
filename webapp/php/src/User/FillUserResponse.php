<?php

declare(strict_types=1);

namespace IsuPipe\User;

use PDO;
use RuntimeException;

trait FillUserResponse
{
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
            throw new RuntimeException();
        }
        $themeModel = ThemeModel::fromRow($row);

        return new User(
            id: $userModel->id,
            name: $userModel->name,
            displayName: $userModel->displayName,
            description: $userModel->description,
            theme: new Theme(
                id: $themeModel->id,
                darkMode: $themeModel->darkMode,
            ),
        );
    }
}
