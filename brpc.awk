#!/bin/awk

# rpc Echo(EchoRequest) returns (EchoResponse);

function extract_method_name(s) {
    return substr(s, 0, index(s, "(") - 1);
}

function extract_request_name(s) {
    start = index(s, "(") + 1;
    end = index(s, ")");
    len = end - start;
    return substr(s, start, len);
}

function extract_response_name(s) {
    start = index(s, "(") + 1;
    end = index(s, ")");
    len = end - start;
    return substr(s, start, len);
}

BEGIN {
    cnt = 0;
}

{
    if ($1 == "service") {    
        service_name = $2;
    } else if ($1 == "rpc") {
        method_names[cnt] = extract_method_name($2);
        req_names[cnt] = extract_request_name($2);
        resp_names[cnt] = extract_response_name($4);
        cnt++;
    } else if ($3 == "};") {
        # EOF
    }
}

END {
    printf "type %s_Stub interface {\n", service_name;
    for (i = 0; i < cnt; i++) {
        printf "\tfunc %s(*Controller, \*%s, \*%s, *RPCDone);\n", method_names[i], req_names[i], resp_names[i];
    }
    printf "}\n";

    printf "type %s interface {\n", service_name;
    for (i = 0; i < cnt; i++) {
        printf "\tfunc %s(*Controller, \*%s, \*%s, *RPCDone);\n", method_names[i], req_names[i], resp_names[i];
    }
    printf "}\n";
}
