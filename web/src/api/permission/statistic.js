import service from "@/utils/request"
import {USER_API} from "@/utils/requestConstants"

// GetClientStatisticData
export const getData = (params) => {
    return service({
        url: `${USER_API}/statistic/client`,
        method: "get",
        params,
    });
};
