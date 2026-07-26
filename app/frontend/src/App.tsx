import { Route, Routes } from "react-router-dom"
import { Header } from "./components/Header"
import { Studio } from "./components/Studio"
import { CensusPage } from "./components/CensusPage"

export default function App() {
  return (
    <div className="flex h-screen flex-col overflow-hidden">
      <Header />
      <main className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <Routes>
          <Route path="/" element={<Studio />} />
          <Route path="/census" element={<div className="overflow-y-auto"><CensusPage /></div>} />
        </Routes>
      </main>
    </div>
  )
}
