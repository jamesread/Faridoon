<?php

namespace faridoon;

use libAllure\Form;
use libAllure\ElementInput;
use libAllure\ElementTextbox;
use libAllure\ElementCheckbox;
use libAllure\ElementSelect;
use libAllure\Session;

class FormQuote extends Form
{
    private $isEdit;

    public function __construct(array $quote = null)
    {
        parent::__construct('add', 'Add Quote');
        $this->setFullyQualifiedElementNames(false);

        $this->isEdit = is_array($quote);

        if ($this->isEdit) {
            $this->addElementHidden('id', $quote['id']);
            $content = $quote['content'];
            $this->setTitle('Editing...');
        } else {
            $content = 'Your quote here.';
        }

		$tb = new ElementTextbox('content', 'Content', stripslashes($content), 'Note: Usernames are automatically highlighted. Timestamps are automatically stripped.');
		$tb->rows = 8;
		$tb->cols = 20;
		$this->addElement($tb);

        global $cfg;
        if ($cfg->getBool('ENABLE_SYNTAX_HIGHLIGHTING')) {
            $this->addSyntaxHighlighting();
        } else {
            $this->addElementHidden('syntaxHighlighting', '');
        }

        $this->addElement(new ElementCheckbox('fixDiscordLinebreaks', 'Fix Discord linebreaks and remove timestamps?', false, 'If you paste a quote from Discord, it may have a lot of newlines around the username. This should fix that.'));

        if ($this->isEdit) {
            $this->addDefaultButtons('Save');
        } else {
            $this->addDefaultButtons('Add');
        }
    }

    private function addSyntaxHighlighting()
    {
        $el = $this->addElement(new ElementSelect('syntaxHighlighting', 'Syntax highlighting for code?', false, 'Is this quote mostly code? If so, it will have pretty formatting applied and usernames will not be highlighted.'));
        $el->addOption('Nope', '');
        $el->addOption('C#', 'csharp');
        $el->addOption('Javascript', 'javascript');
        $el->addOption('PHP', 'php');
        $el->addOption('Java', 'java');
        $el->addOption('Python', 'python');

        $this->addElement($el);
    }

    private function fixDiscordLinebreaks($content)
    {
        $content = preg_replace('#\[\n(?<date>\d\d:\d\d)\n\]\n(?<username>[\w\d_]+)\n:\n#i', "\\2: ", $content, -1, $count);

        return $content;
    }

    public function process()
    {
        $content = $this->getElementValue('content');
        $content = preg_replace('#\r\n#', "\n", $content);

        if ($this->getElementValue('fixDiscordLinebreaks')) {
            $content = $this->fixDiscordLinebreaks($content);
        }

        if ($this->isEdit) {
            $this->processEdit($content);
        } else {
            $this->processAdd($content);
        }
    }

    public function processEdit($content)
    {
        global $db;
        global $cfg;

        if ($cfg->getBool('ENABLE_SYNTAX_HIGHLIGHTING')) {
            $syntaxHighlighting = $this->getElementValue('syntaxHighlighting');
        } else {
            $syntaxHighlighting = '';
        }

        $sql = 'UPDATE quotes SET content = :content, syntaxHighlighting = :syntaxHighlighting WHERE id = :id ';
        $stmt = $db->prepare($sql);
        $stmt->bindValue(':content', $content);
        $stmt->bindValue(':syntaxHighlighting', $syntaxHighlighting);
        $stmt->bindValue(':id', $this->getElementValue('id'));
        $stmt->execute();
    }

    public function processAdd($content)
    {
        global $db;

        $sql = 'INSERT INTO quotes (content, created, approval) VALUES (:content, now(), :approval) ';
        $stmt = $db->prepare($sql);
        $stmt->bindValue(':content', $content);

        if (Session::isLoggedIn()) {
            $stmt->bindValue(':approval', Session::getUser()->hasPriv('BYPASS_APPROVAL') ? 1 : 0);
        } else {
            $stmt->bindValue(':approval', 0);
        }

        $stmt->execute();
    }
}
