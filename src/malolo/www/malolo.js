fxml_url = 'http://jamesnewman2015:9b73fe05a1917fb7c621d864dc119321c269fe8c@flightxml.flightaware.com/json/FlightXML2/';
flight_aware_success=true;



function SeeRoute() {
    document.getElementById('line1').innerHTML="Origin: DCA ----> Destination: ORD";
    document.getElementById('line2').innerHTML="Altitude: 35,000 ft. (Inflight Wi-Fi Approved Altitude)";
}

function SeeNetwork() {
    document.getElementById('line1').innerHTML="Your Latency: 500ms";
    document.getElementById('line2').innerHTML="Your Bandwidth: 0.5Mbps";
}

//When user inputs flight information, bar graph reflects labels with new data
function SeeMyFlight() {
    //Hide alert boxes, and show loader 
    document.getElementById("waiting").hidden=false;
    document.getElementById("flight_success").hidden=true;
    document.getElementById("flight_summary").hidden=true;
    // document.getElementById("startTest").visibility=false;
	console.log("MyFlight function called")


    //Process Incoming Data
    var flight=(document.getElementById("ident_text").value).toUpperCase();
    console.log(flight);
    lat_dataset[4].label=flight;
    bw_dataset[4].label=flight;
    loss_dataset[4].label=flight;
    
    //Update Changes
    change(lat_dataset);

    //Wait appropriate amount of time for loader to disappear
    setTimeout(function () {
        //Add alert boxes based on whether FlightAware was successful in querying data
        if (flight_aware_success) {
            document.getElementById("flight_success").hidden=false;
            document.getElementById("flight_summary").hidden=false;
        }
        else {
            document.getElementById("flight_failure").hidden=false;
        }
        document.getElementById("waiting").hidden=true;
        //Visualize changes (By changing which graph appears)
        change(bw_dataset);
        // document.getElementById("startTest").hidden=false;

        //How long should Malolo wait before it loads (set to half a second for testing purposes)
    }, 500);
	
   

}


//When measurements come in from Go coroutines, adjust values accordingly, or post to servers
function changeLat(val) {
    lat_dataset[4].value=val;
}
function changeBW(val) {
    lat_dataset[4].value=val;
}
function changeLoss(val) {
    lat_dataset[4].label=val;
}

//When measurements come in we want to post, save as request_data variable and then call this function
function post_data() {
    $.post('http://hinckley.cs.northwestern.edu/~rbi054/post.php/',  
        { data: request_data }, 
        function(data, status, xhr) {
        
            $('p').append('status: ' + status + ', data: ' + data);

        }).done(function() { console.log('Request done!'); })
        .fail(function(jqxhr, settings, ex) { console.log('failed, ' + ex); });
    }

//At some point, add valid flight checking (currently, Javascript is browser is blocking request for security reasons)
    //More info: https://stackoverflow.com/questions/20035101/why-does-my-javascript-get-a-no-access-control-allow-origin-header-is-present
    //  $.post("http://hinckley.cs.northwestern.edu/~jtn609/flightaware/get_inflightinfo.php",
    //   { ident: flight },
    //   function(data, status){
    //      console.log(data);
    //      //Find Origin and Destination
    //      console.log(data.toString());
    //      string_data=data.toString();
    //      var org1=parseInt(data.search("origin"));
    //      console.log(data[org1+10]+data[org1+11]+data[org1+12]);
    //      var dest1=data.search("destination");
    //      console.log(data[dest1+15]+data[dest1+16]+data[dest1+17]);
    //      oldorigin=String(data[org1+9]+data[org1+10]+data[org1+11]+data[org1+12]);
    //      olddestination=String(data[dest1+14]+data[dest1+15]+data[dest1+16]+data[dest1+17]);
    // });
// }

//Add Triggerable Buttons
document.addEventListener('DOMContentLoaded', function(){
    console.log("hello");
    document.getElementById('startTest').onclick=SeeMyFlight; 
    document.getElementById('route_button').onclick=SeeRoute;
    document.getElementById('network_button').onclick=SeeNetwork;
    }, false);

//Allow user to press enter when entering flight
window.onload = function () { 
    var input = document.getElementById("ident_text");
    // Execute a function when the user releases a key on the keyboard
    input.addEventListener("keyup", function(event) {
      // Cancel the default action, if needed
      event.preventDefault();
      // Number 13 is the "Enter" key on the keyboard
      if (event.keyCode === 13) {
        // Trigger the button element with a click
        console.log("Enter key fired");
        document.getElementById("startTest").click();
      }
    });        
}

