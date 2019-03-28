import axios from 'axios';
import { LOADING } from '../Map/actions'

export const SET_BUSINESS = 'set_business';
export const SET_ITEMS = 'set_items';
export const SHOW_MODAL = 'show_modal';

const url = process.env.REACT_APP_BACKEND_URL;

export const getItems = place => async dispatch => {
    console.log("url:", url)
    dispatch({ type: LOADING, payload: true })
    dispatch({ type: SET_BUSINESS, payload: place })
    const items = await axios.get(`${url}/items/${place.r_id}`);
    dispatch({
        type: SET_ITEMS,
        payload: items.data
    })
    dispatch({ type: LOADING, payload: false })
}

export const setBusiness = business => {
    return {
        type: SET_BUSINESS,
        payload: business
    }
}

export const setShowModal = (value) => {
    return {
        type: SHOW_MODAL,
        payload: value
    }
}