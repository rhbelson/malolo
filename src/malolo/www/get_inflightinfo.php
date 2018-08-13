<?php
$ident = htmlspecialchars($_POST["ident"]);
$options = array(
    'trace' => true,
    'exceptions' => 0,
    'login' => 'jamesnewman2015',
    'password' => '9b73fe05a1917fb7c621d864dc119321c269fe8c',
);
$client = new SoapClient('http://flightxml.flightaware.com/soap/FlightXML2/wsdl', $options);
$params = array("ident" => $ident);
$result = $client->InFlightInfo($params);
echo json_encode((array)$result);
?>