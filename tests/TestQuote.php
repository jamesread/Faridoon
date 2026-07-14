<?php

use PHPUnit\Framework\TestCase;

class TestQuote extends TestCase
{
    public function testUnmarshalFromDatabaseSetsFields()
    {
        $quote = new faridoon\Quote();
        $quote->unmarshalFromDatabase([
            'id' => 42,
            'created' => '2024-01-02 03:04:05',
            'voteCount' => 7,
            'approved' => 1,
            'content' => "alice: hello\nbob: hi",
        ]);

        $this->assertEquals(42, $quote->id);
        $this->assertEquals('2024-01-02 03:04:05', $quote->created);
        $this->assertEquals(7, $quote->voteCount);
        $this->assertEquals(1, $quote->approved);
        $this->assertCount(2, $quote->lines);
        $this->assertEquals('alice', $quote->lines[0]['username']);
        $this->assertEquals('bob', $quote->lines[1]['username']);
    }

    public function testIrcModePrefixes()
    {
        $text = <<<EOQ
@ops: status
+voice: hi
~owner: hello
&admin: there
EOQ;

        $quote = new faridoon\Quote();
        $quote->unmarshalFromText($text);

        $this->assertEquals('ops', $quote->lines[0]['username']);
        $this->assertEquals('voice', $quote->lines[1]['username']);
        $this->assertEquals('owner', $quote->lines[2]['username']);
        $this->assertEquals('admin', $quote->lines[3]['username']);
    }

    public function testUsernameColorsAreStableAndUnique()
    {
        $text = <<<EOQ
alice: one
bob: two
alice: three
EOQ;

        $quote = new faridoon\Quote();
        $quote->unmarshalFromText($text);

        $this->assertEquals($quote->lines[0]['usernameColor'], $quote->lines[2]['usernameColor']);
        $this->assertNotEquals($quote->lines[0]['usernameColor'], $quote->lines[1]['usernameColor']);
    }

    public function testContinuationLinesHaveNoUsername()
    {
        $text = <<<EOQ
alice: hello
this is a continuation
EOQ;

        $quote = new faridoon\Quote();
        $quote->unmarshalFromText($text);

        $this->assertEquals('alice', $quote->lines[0]['username']);
        $this->assertNull($quote->lines[1]['username']);
        $this->assertEquals('this is a continuation', $quote->lines[1]['content']);
    }

    public function testEmptyMessageAfterUsernameIsIgnored()
    {
        $text = "alice: \nbob: hello";

        $quote = new faridoon\Quote();
        $quote->unmarshalFromText($text);

        $this->assertNull($quote->lines[0]['username']);
        $this->assertEquals('bob', $quote->lines[1]['username']);
        $this->assertEquals('hello', $quote->lines[1]['content']);
    }

    public function testHtmlIsEscapedInParsedLines()
    {
        $text = 'alice: <script>alert(1)</script>';

        $quote = new faridoon\Quote();
        $quote->unmarshalFromText($text);

        $this->assertEquals('alice', $quote->lines[0]['username']);
        $this->assertStringNotContainsString('<script>', $quote->lines[0]['content']);
        $this->assertStringContainsString('&lt;script&gt;', $quote->lines[0]['content']);
    }
}
