<?php

namespace faridoon;

class ElementTextbox extends \libAllure\ElementTextbox
{
    public $placeholder = '';

    public function render()
    {
        if ($this->value == null) {
            $this->value = '';
        }

        $placeholder = '';
        if ($this->placeholder !== '') {
            $placeholder = sprintf(' placeholder = "%s"', htmlspecialchars($this->placeholder, ENT_QUOTES));
        }

        return sprintf(
            '<textarea id = "%s" name = "%s" rows = "%s" cols = "%s"%s>%s</textarea>',
            $this->name,
            $this->name,
            $this->rows,
            $this->cols,
            $placeholder,
            $this->value
        );
    }
}
