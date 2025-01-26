import { Container, Stack } from "@chakra-ui/react";
import Navbar from "./components/Navbar";
import { Route, Routes } from 'react-router';
import { Todos } from "./pages/Todos";
import { Profile } from "./pages/Profile";

export const BASE_URL = import.meta.env.MODE === "development" ? "http://localhost:4000/api" : "/api";

function App() {
	return (
		<Stack h='100vh'>
			<Navbar />
			<Container>
					<Routes>
						<Route path='/' element={<Todos />} />
						<Route path='/profile' element={<Profile />} />
					</Routes>
			</Container>
		</Stack>
	);
}

export default App;