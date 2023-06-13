import React from 'react';
import { useDispatch, useSelector } from 'react-redux';
import ImgSrc from './assets/images/beautiful.jpg';
import { decrement, decrementBy, increment, incrementBy } from './redux/counter';

export default function App() {
    const { count } = useSelector((state) => state.counter);
    const dispatch = useDispatch();
    return (
        <div className="App">
            <h1> The count is: {count}</h1>
            <button onClick={() => dispatch(increment())}>increment</button>
            <button onClick={() => dispatch(decrement())}>decrement</button>
            <button onClick={() => dispatch(incrementBy(33))}>Increment by 33</button>
            <button onClick={() => dispatch(decrementBy(33))}>Decrement by 33</button>
            <br />
            <br />
            <img src={ImgSrc} />
        </div>
    );
}
