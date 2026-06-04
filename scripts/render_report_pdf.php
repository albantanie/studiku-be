<?php

require __DIR__ . '/../php_vendor/autoload.php';

use Dompdf\Dompdf;
use Dompdf\Options;

$html = stream_get_contents(STDIN);
if ($html === false || trim($html) === '') {
    fwrite(STDERR, "HTML input is empty\n");
    exit(1);
}

$options = new Options();
$options->set('isRemoteEnabled', true);
$options->set('isHtml5ParserEnabled', true);
$options->set('defaultFont', 'Arial');

$dompdf = new Dompdf($options);
$dompdf->loadHtml($html, 'UTF-8');
$dompdf->setPaper('A4', 'portrait');
$dompdf->render();

echo $dompdf->output();
