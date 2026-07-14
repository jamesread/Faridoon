<?php

use PHPUnit\Framework\TestCase;

class TestSchemaMigrations extends TestCase
{
    private function baseSql(): string
    {
        return file_get_contents(__DIR__ . '/../database/migrations/0.base.sql');
    }

    private function charsetMigration(): string
    {
        return file_get_contents(__DIR__ . '/../database/migrations/3.charset-utf8mb4.sql');
    }

    public function testBaseSchemaUsesUtf8mb4()
    {
        $sql = $this->baseSql();

        $this->assertStringNotContainsString('CHARSET=latin1', $sql);
        $this->assertStringContainsString('DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci', $sql);
        $this->assertStringContainsString('character_set_client = utf8mb4', $sql);
    }

    public function testBaseSchemaDoesNotSeedAutoIncrement()
    {
        $sql = $this->baseSql();

        $this->assertDoesNotMatchRegularExpression('/AUTO_INCREMENT=\d+/', $sql);
    }

    public function testBaseSchemaUsesInnoDb()
    {
        $sql = $this->baseSql();

        $this->assertStringContainsString('ENGINE=InnoDB', $sql);
        $this->assertStringNotContainsString('ENGINE=MyISAM', $sql);
    }

    public function testCharsetUpgradeMigrationConvertsAllTables()
    {
        $sql = $this->charsetMigration();
        $tables = [
            'group_memberships',
            'groups',
            'permissions',
            'privileges_g',
            'privileges_u',
            'quotes',
            'users',
            'votes',
        ];

        foreach ($tables as $table) {
            $this->assertStringContainsString(
                "ALTER TABLE `{$table}` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;",
                $sql
            );
        }
    }

    public function testApplicationRequiresCharsetMigration()
    {
        $common = file_get_contents(__DIR__ . '/../src/includes/common.php');

        $this->assertStringContainsString("requireDatabaseVersion('3.charset-utf8mb4.sql');", $common);
    }
}
